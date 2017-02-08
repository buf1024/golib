package net

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	EventNone = iota
	EventNewConnected
	EventError
	EventNewData
	EventTimer
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
	Net    *SimpleNet
	listen *Listener

	id      int
	status  int64
	conn    net.Conn
	msgChan chan []byte

	LocalAddr  string
	RemoteAddr string
	Proto      IProto // 为了实现多种proto
	UpTime     time.Time
	UserData   interface{}
}
type Listener struct {
	Net *SimpleNet

	id     int
	status int64
	listen net.Listener
	conns  []*Connection

	lockClient sync.Locker

	UserData interface{}
}

type SimpleNet struct {
	proto  IProto
	events chan *ConnEvent

	connClient []*Connection
	connServer []*Listener

	lockServer sync.Locker
	lockClient sync.Locker

	UserData interface{}
}

func (c *Connection) ID() string {
	id := fmt.Sprintf("%d", c.id)
	if c.listen != nil {
		id = fmt.Sprintf("%d-%s", c.listen.id, id)
	}
	return id
}

func (l *Listener) ID() string {
	return fmt.Sprintf("%d", l.id)
}

type IProto interface {
	HeadLen() int
	BodyLen(head []byte) ([]byte, int)
	Parse(head []byte, body []byte) (interface{}, error)
	Serialize(data interface{}) ([]byte, error)
}

func NewSimpleNet(proto IProto, userData interface{}) *SimpleNet {
	n := &SimpleNet{
		proto:      proto,
		events:     make(chan *ConnEvent, 1024),
		connClient: make([]*Connection, 1),
		connServer: make([]*Listener, 1),
		lockServer: &sync.Mutex{},
		lockClient: &sync.Mutex{},
		UserData:   userData,
	}

	return n
}

func SimpleNetDestroy(c *SimpleNet) {

}

func (c *SimpleNet) syncAddListen(listen *Listener) int {
	c.lockServer.Lock()
	defer c.lockServer.Unlock()

	c.connServer = append(c.connServer, listen)

	return len(c.connServer) - 1

}
func (c *SimpleNet) syncAddClient(conn *Connection) int {
	c.lockClient.Lock()
	defer c.lockClient.Unlock()

	c.connClient = append(c.connClient, conn)

	return len(c.connClient) - 1
}
func (c *SimpleNet) syncAddListenClient(listen *Listener, conn *Connection) int {
	listen.lockClient.Lock()
	defer listen.lockClient.Unlock()

	listen.conns = append(listen.conns, conn)

	return len(listen.conns) - 1
}

func (c *SimpleNet) checkConnErr(err error, conn *Connection) error {
	if err != nil && conn.status == StatusConnected {
		close(conn.msgChan)
		conn.conn.Close()
		conn.status = StatusBroken

		// emit EventError
		event := &ConnEvent{
			EventType: EventError,
			Conn:      conn,
			Data:      err,
		}
		c.events <- event
	}
	return err
}
func (c *SimpleNet) handleRead(conn *Connection) {

	for {
		headlen := 0
		if conn.Proto != nil {
			headlen = conn.Proto.HeadLen()
		}
		if headlen <= 0 {
			buf := make([]byte, 1)
			_, err := conn.conn.Read(buf)
			if err = c.checkConnErr(err, conn); err != nil {
				return
			}
			// emit EventNewData
			event := &ConnEvent{
				EventType: EventNewData,
				Conn:      conn,
				Data:      buf,
			}
			c.events <- event

		} else {
			head := make([]byte, headlen)
			_, err := conn.conn.Read(head)
			if err = c.checkConnErr(err, conn); err != nil {
				return
			}
			head, bodylen := conn.Proto.BodyLen(head)
			body := make([]byte, bodylen)
			if err = c.checkConnErr(err, conn); err != nil {
				return
			}
			data, err := conn.Proto.Parse(head, body)
			if err = c.checkConnErr(err, conn); err != nil {
				return
			}
			// emit EventNewData
			event := &ConnEvent{
				EventType: EventNewData,
				Conn:      conn,
				Data:      data,
			}
			c.events <- event
		}
	}
}

func (c *SimpleNet) handleWrite(conn *Connection) {
	for {
		select {
		case msg, ok := <-conn.msgChan:
			{
				if !ok {
					return
				}
				_, err := conn.conn.Write(msg)
				if err = c.checkConnErr(err, conn); err != nil {
					return
				}
			}
		}
	}
}

func (c *SimpleNet) listening(l *Listener) {
	for {
		newconn, err := l.listen.Accept()
		if err != nil {
			fmt.Printf("accept failed, err = %s\n", err)
			continue
		}

		conn := &Connection{
			Net:        l.Net,
			listen:     l,
			status:     StatusConnected,
			conn:       newconn,
			msgChan:    make(chan []byte, 1024),
			LocalAddr:  newconn.LocalAddr().String(),
			RemoteAddr: newconn.RemoteAddr().String(),
			Proto:      l.Net.proto,
			UpTime:     time.Now(),
		}
		conn.id = c.syncAddListenClient(l, conn)

		// emit EventNewConnected
		event := &ConnEvent{
			EventType: EventNewConnected,
			Conn:      conn,
		}
		c.events <- event

		go c.handleRead(conn)
		go c.handleWrite(conn)

	}
}

func (c *SimpleNet) Listen(addr string) (*Listener, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	l := &Listener{
		Net:        c,
		status:     StatusListenning,
		listen:     listen,
		conns:      make([]*Connection, 1),
		lockClient: &sync.Mutex{},
	}
	l.id = c.syncAddListen(l)

	go c.listening(l)

	return l, nil
}
func (c *SimpleNet) Connect(host string) (*Connection, error) {
	n, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		Net:        c,
		status:     StatusConnected,
		conn:       n,
		msgChan:    make(chan []byte, 1024),
		LocalAddr:  n.LocalAddr().String(),
		RemoteAddr: n.RemoteAddr().String(),
		Proto:      c.proto,
		UpTime:     time.Now(),
	}
	conn.id = c.syncAddClient(conn)

	go c.handleRead(conn)
	go c.handleWrite(conn)

	return conn, nil
}

func (c *SimpleNet) PollEvent() (*ConnEvent, error) {
	select {
	case event, ok := <-c.events:
		{
			if !ok {
				return nil, fmt.Errorf("SimpleNet destroyed")
			}
			return event, nil
		}
	}
}

func (c *SimpleNet) SendData(conn *Connection, data interface{}) error {
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
