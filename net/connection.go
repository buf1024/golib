package net

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	EventNone = iota
	EventNewConnected
	EventError
	EventNewData
	EventTimeout
)

const (
	StatusNone = iota
	StatusListenning
	StatusConnected
	StatusBroken
)

type ConnEvent struct {
	EventType int64
	Conn      *Connection
	Data      interface{}
}

type Connection struct {
	net    *SimpleNet
	listen *Listener

	id      int64
	status  int64
	conn    net.Conn
	msgChan chan []byte

	localAddr  string
	remoteAddr string
	upTime     time.Time

	Proto    IProto // 为了实现多种proto
	UserData interface{}
}

func (c *Connection) Net() *SimpleNet {
	return c.net
}
func (c *Connection) ID() int64 {
	return c.id
}

func (c *Connection) Status() int64 {
	return c.status
}

func (c *Connection) LocalAddress() string {
	return c.localAddr
}

func (c *Connection) RemoteAddress() string {
	return c.remoteAddr
}

type Listener struct {
	net *SimpleNet

	id     int64
	status int64
	listen net.Listener
	conns  []*Connection

	lockClient sync.Locker

	Proto    IProto
	UserData interface{}
}

func (l *Listener) ID() int64 {
	return l.id
}
func (l *Listener) Net() *SimpleNet {
	return l.net
}
func (l *Listener) LocalAddress() string {
	return l.listen.Addr().String()
}

type SimpleNet struct {
	events chan *ConnEvent

	connClient []*Connection
	connServer []*Listener

	lockServer sync.Locker
	lockClient sync.Locker

	nextid int64

	UserData interface{}
}

type IProto interface {
	FilterAccept(conn *Connection) bool
	HeadLen() int
	BodyLen(head []byte) (interface{}, int, error)
	Parse(head interface{}, body []byte) (interface{}, error)
	Serialize(data interface{}) ([]byte, error)
}

// NewSimpleNet 创建
func NewSimpleNet() *SimpleNet {
	n := &SimpleNet{
		events:     make(chan *ConnEvent, 1024),
		connClient: make([]*Connection, 1),
		connServer: make([]*Listener, 1),
		lockServer: &sync.Mutex{},
		lockClient: &sync.Mutex{},
	}

	return n
}

func SimpleNetDestroy(n *SimpleNet) {
	close(n.events)
	for _, v := range n.connClient {
		n.CloseConn(v)
	}

	for _, v := range n.connServer {
		n.CloseListen(v)
	}
}

func (n *SimpleNet) syncAddListen(listen *Listener) {
	n.lockServer.Lock()
	defer n.lockServer.Unlock()

	n.connServer = append(n.connServer, listen)

}
func (n *SimpleNet) syncDelListen(listen *Listener) {
	n.lockServer.Lock()
	defer n.lockServer.Unlock()

	for i, v := range n.connServer {
		if v == listen {
			if i == len(n.connServer)-1 {
				n.connServer = n.connServer[:i]
			} else {
				n.connServer = append(n.connServer[:i], n.connServer[i:]...)
			}
		}
	}

}

func (n *SimpleNet) syncAddClient(conn *Connection) {
	var connQueue []*Connection

	connQueue = n.connClient
	lock := n.lockClient
	if conn.listen != nil {
		connQueue = conn.listen.conns
		lock = conn.listen.lockClient
	}

	lock.Lock()
	defer lock.Unlock()

	connQueue = append(connQueue, conn)

	if conn.listen != nil {
		conn.listen.conns = connQueue
	} else {
		n.connClient = connQueue
	}
}
func (n *SimpleNet) syncDelClient(conn *Connection) {
	var connQueue []*Connection

	connQueue = n.connClient
	lock := n.lockClient
	if conn.listen != nil {
		connQueue = conn.listen.conns
		lock = conn.listen.lockClient
	}

	lock.Lock()
	defer lock.Unlock()

	var del bool
	for i, v := range connQueue {
		if v == conn {
			if i == len(connQueue)-1 {
				connQueue = connQueue[:i]
			} else {
				connQueue = append(connQueue[:i], connQueue[i:]...)
			}
			del = true
		}
	}

	if del {
		if conn.listen != nil {
			conn.listen.conns = connQueue
		} else {
			n.connClient = connQueue
		}
	}
}

func (n *SimpleNet) checkConnErr(err error, conn *Connection) error {
	if err != nil && conn.status == StatusConnected {
		close(conn.msgChan)
		conn.conn.Close()
		conn.status = StatusBroken

		n.syncDelClient(conn)

		// emit EventError
		event := &ConnEvent{
			EventType: EventError,
			Conn:      conn,
			Data:      err,
		}
		n.events <- event
	}
	return err
}
func (n *SimpleNet) handleRead(conn *Connection) {

	for {
		headlen := 0
		if conn.Proto != nil {
			headlen = conn.Proto.HeadLen()
		}
		if headlen <= 0 {
			buf := make([]byte, 1)
			_, err := conn.conn.Read(buf)
			if err = n.checkConnErr(err, conn); err != nil {
				return
			}
			// emit EventNewData
			event := &ConnEvent{
				EventType: EventNewData,
				Conn:      conn,
				Data:      buf,
			}
			n.events <- event

		} else {
			head := make([]byte, headlen)
			_, err := conn.conn.Read(head)
			if err = n.checkConnErr(err, conn); err != nil {
				return
			}
			headmsg, bodylen, err := conn.Proto.BodyLen(head)
			if err = n.checkConnErr(err, conn); err != nil {
				return
			}
			body := make([]byte, bodylen)
			data, err := conn.Proto.Parse(headmsg, body)
			if err = n.checkConnErr(err, conn); err != nil {
				return
			}
			// emit EventNewData
			event := &ConnEvent{
				EventType: EventNewData,
				Conn:      conn,
				Data:      data,
			}
			n.events <- event
		}
		conn.upTime = time.Now()
	}
}

func (n *SimpleNet) handleWrite(conn *Connection) {
	for {
		select {
		case msg, ok := <-conn.msgChan:
			{
				if !ok {
					return
				}
				_, err := conn.conn.Write(msg)
				if err = n.checkConnErr(err, conn); err != nil {
					return
				}
				conn.upTime = time.Now()
			}
		}
	}
}

func (n *SimpleNet) listening(l *Listener) {
	for {
		newconn, err := l.listen.Accept()
		if err != nil {
			fmt.Printf("accept failed, err = %s\n", err)
			if l.status != StatusListenning {
				break
			}
			continue
		}

		conn := &Connection{
			net:        l.net,
			listen:     l,
			id:         atomic.AddInt64(&n.nextid, 1),
			status:     StatusConnected,
			conn:       newconn,
			msgChan:    make(chan []byte, 1024),
			localAddr:  newconn.LocalAddr().String(),
			remoteAddr: newconn.RemoteAddr().String(),
			Proto:      l.Proto,
			upTime:     time.Now(),
		}

		if conn.Proto != nil {
			if !conn.Proto.FilterAccept(conn) {
				continue
			}
		}

		n.syncAddClient(conn)

		// emit EventNewConnected
		event := &ConnEvent{
			EventType: EventNewConnected,
			Conn:      conn,
		}
		n.events <- event

		go n.handleRead(conn)
		go n.handleWrite(conn)

	}
}

// Listen 监听网络 addr 为监听地址
func (n *SimpleNet) Listen(addr string) (*Listener, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		net: n,

		id:         atomic.AddInt64(&n.nextid, 1),
		status:     StatusListenning,
		listen:     listen,
		conns:      make([]*Connection, 1),
		lockClient: &sync.Mutex{},
	}
	n.syncAddListen(l)

	go n.listening(l)

	return l, nil
}

// Connect 连接服务器器
func (n *SimpleNet) Connect(addr string) (*Connection, error) {
	newconn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		net:        n,
		id:         atomic.AddInt64(&n.nextid, 1),
		status:     StatusConnected,
		conn:       newconn,
		msgChan:    make(chan []byte, 1024),
		localAddr:  newconn.LocalAddr().String(),
		remoteAddr: newconn.RemoteAddr().String(),
		upTime:     time.Now(),
	}
	n.syncAddClient(conn)

	go n.handleRead(conn)
	go n.handleWrite(conn)

	return conn, nil
}

// PollEvent 事件轮询
func (n *SimpleNet) PollEvent(timeout int) (*ConnEvent, error) {
	t := time.After(time.Millisecond * (time.Duration)(timeout))
	select {
	case event, ok := <-n.events:
		{
			if !ok {
				return nil, fmt.Errorf("SimpleNet destroyed")
			}
			return event, nil
		}
	case <-t:
		{
			evt := &ConnEvent{
				EventType: EventTimeout,
			}
			return evt, nil
		}
	}
}

// SendData 向connection发送数据，如果connection不支持，data为[]byte
func (n *SimpleNet) SendData(conn *Connection, data interface{}) error {
	if conn.Proto == nil {
		msg, ok := (data).([]byte)
		if !ok {
			return fmt.Errorf("unexpect data type")
		}
		conn.msgChan <- msg
	} else {
		msg, err := conn.Proto.Serialize(data)
		if err != nil {
			return err
		}
		conn.msgChan <- msg
	}
	return nil
}

// CloseConn 关闭连接
func (n *SimpleNet) CloseConn(conn *Connection) error {
	conn.status = StatusBroken
	close(conn.msgChan)
	conn.conn.Close()

	n.syncDelClient(conn)

	return nil
}

// CloseListen 关闭服务器
func (n *SimpleNet) CloseListen(listen *Listener) error {
	for _, v := range listen.conns {
		n.CloseConn(v)
	}
	listen.status = StatusBroken
	listen.listen.Close()

	return nil
}
