package main

import (
	"fmt"
	"os"

	"time"

	"github.com/buf1024/golib/exam/pb"
	mynet "github.com/buf1024/golib/net"
	"github.com/golang/protobuf/proto"
)

type pbserver struct {
	n      *mynet.SimpleNet
	listen *mynet.Listener
	conns  []*mynet.Connection
	proto  *pb.PbServerProto
}

func (s *pbserver) handler(conn *mynet.Connection, msg *pb.PbProto) {
	fmt.Printf("REQ:\n%s\n", s.proto.Debug(msg))

	switch msg.H.Command {
	case pb.CMDHeartBeatReq:
		{
			req := msg.B.(*pb.HeartBeatReq)

			rsp := &pb.PbProto{}
			rsp.H.Command = pb.CMDHeartBeatRsp
			// beartbeat
			p, err := s.proto.GetMessage(rsp.H.Command)
			if err != nil {
				fmt.Printf("gen hearbeat rsp failed, err = %s\n", err)
				return
			}
			pheart := p.(*pb.HeartBeatRsp)
			pheart.SID = proto.String(req.GetSID())

			rsp.B = pheart
			fmt.Printf("RSP：\n%s\n", s.proto.Debug(rsp))

			err = s.n.SendData(conn, rsp)

			if err != nil {
				fmt.Printf("send hearbeat rsp, err = %s\n", err)
				return
			}
		}
	case pb.CMDHeartBeatRsp:
		{

		}
	case pb.CMDBizReq:
		{
			req := msg.B.(*pb.BizReq)

			rsp := &pb.PbProto{}
			rsp.H.Command = pb.CMDBizRsp
			// BIZRSP
			p, err := s.proto.GetMessage(rsp.H.Command)
			if err != nil {
				fmt.Printf("gen hearbeat rsp failed, err = %s\n", err)
				return
			}
			pbiz := p.(*pb.BizRsp)
			pbiz.SID = proto.String(req.GetSID())
			pbiz.RetCode = proto.Int32(9999)

			rsp.B = pbiz
			fmt.Printf("RSP：\n%s\n", s.proto.Debug(rsp))

			err = s.n.SendData(conn, rsp)

			if err != nil {
				fmt.Printf("send biz rsp, err = %s\n", err)
				return
			}
		}
	default:
		fmt.Printf("unknow command 0x%x\n", msg.H.Command)
	}
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

func (s *pbserver) hearbeat() {
END:
	for {
		t := time.After(time.Second * 5)
		select {
		case <-t:
			{
				if s.conns == nil {
					return
				}
				if len(s.conns) == 0 {
					continue END
				}
				n := time.Now()
				for _, v := range s.conns {
					if v.Status() != mynet.StatusBroken {
						up := v.UpdateTime()
						diff := n.Unix() - up.Unix()
						fmt.Printf("now = %d, up = %d diff = %d\n",
							n.Unix(), up.Unix(), diff)
						if diff > 10 {
							m := &pb.PbProto{}
							m.H.Command = pb.CMDHeartBeatReq
							// beartbeat
							p, err := s.proto.GetMessage(m.H.Command)
							if err != nil {
								fmt.Printf("gen hearbeat failed, err = %s\n", err)
								continue
							}
							pheart := p.(*pb.HeartBeatReq)
							pheart.SID = proto.String(pb.SID(32))
							m.B = pheart

							fmt.Printf("SEND：\n%s\n", s.proto.Debug(m))
							err = s.n.SendData(v, m)

							if err != nil {
								fmt.Printf("send hearbeat, err = %s\n", err)
								continue
							}
						}

					}
				}
			}

		}
	}

}

func (s *pbserver) server() {
	go s.hearbeat()

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
		case evt.EventType == mynet.EventConnectionError:
			{
				fmt.Printf("event error: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
				s.del(conn)
			}
		case evt.EventType == mynet.EventConnectionClosed:
			{
				fmt.Printf("event close: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
				s.del(conn)
			}
		case evt.EventType == mynet.EventNewConnectionData:
			{
				data := evt.Data.(*pb.PbProto)
				go s.handler(conn, data)
			}
		case evt.EventType == mynet.EventNewConnection:
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
	s.conns = nil
	mynet.SimpleNetDestroy(s.n)

}

func main() {
	s := &pbserver{
		n:     mynet.NewSimpleNet(),
		proto: &pb.PbServerProto{},
	}

	listen, e := s.n.Listen("127.0.0.1:4369", s.proto)
	if e != nil {
		fmt.Printf("listen failed, err=%s\n", e)
		os.Exit(-1)
	}
	s.listen = listen

	s.server()

	s.destroy()
}
