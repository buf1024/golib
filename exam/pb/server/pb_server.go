package main

import (
	"fmt"
	"os"

	"time"

	"github.com/buf1024/golib/exam/pb"
	mynet "github.com/buf1024/golib/net"
)

type pbserver struct {
	n      *mynet.SimpleNet
	listen *mynet.Listener
	conns  []*mynet.Connection
}

func (s *pbserver) handler(msg *pb.PbProto) {
}

func (s *pbserver) add(con *mynet.Connection) {
	s.conns = append(s.conns, con)
}
func (s *pbserver) del(con *mynet.Connection) {
	for i, v := range s.conns {
		if con == v {
			if i == len(s.conns)-1 {
				s.conns = s.conns[:i]
			} else {
				s.conns = append(s.conns[:i], s.conns[i+1:]...)
			}
			break
		}
	}
}

func (s *pbserver) server() {
	n := s.n
	listen := s.listen

	fmt.Printf("server listenning %s\n", listen.LocalAddress())

	for {
		evt, err := n.PollEvent(1000 * 60)
		if err != nil {
			fmt.Printf("poll event error!")
			return
		}
		conn := evt.Conn
		switch {
		case evt.EventType == mynet.EventError:
			{
				fmt.Printf("event error: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
				s.del(conn)
			}
		case evt.EventType == mynet.EventNewData:
			{
				data := evt.Data.(*pb.PbProto)
				go s.handler(data)
			}
		case evt.EventType == mynet.EventNewConnected:
			{
				fmt.Printf("client conneced: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
				s.add(conn)
			}
		case evt.EventType == mynet.EventTimeout:
			{
				fmt.Printf("poll timeout now = %d\n", time.Now().Unix())
			}
		}
	}
}

func (s *pbserver) destroy() {
	mynet.SimpleNetDestroy(s.n)

}

func main() {
	s := &pbserver{
		n:     mynet.NewSimpleNet(),
		conns: make([]*mynet.Connection, 1),
	}

	listen, e := s.n.Listen("127.0.0.1:4369")
	if e != nil {
		fmt.Printf("listen failed, err=%s\n", e)
		os.Exit(-1)
	}
	s.listen = listen

	s.server()

	s.destroy()
}
