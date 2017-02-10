package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"bufio"

	"github.com/buf1024/golib/exam/pb"
	mynet "github.com/buf1024/golib/net"
	"github.com/golang/protobuf/proto"
)

type userData struct {
	w     *sync.WaitGroup
	c     chan string
	proto *pb.PbServerProto
}

func handler(conn *mynet.Connection, msg *pb.PbProto) {
	n := conn.Net()
	data := conn.UserData.(*userData)
	pro := data.proto

	fmt.Printf("REQ:\n%s\n", pro.Debug(msg))

	switch msg.H.Command {
	case pb.CMDHeartBeatReq:
		{
			req := msg.B.(*pb.HeartBeatReq)

			rsp := &pb.PbProto{}
			rsp.H.Command = pb.CMDHeartBeatRsp
			// beartbeat
			p, err := pro.GetMessage(rsp.H.Command)
			if err != nil {
				fmt.Printf("gen hearbeat rsp failed, err = %s\n", err)
				return
			}
			pheart := p.(*pb.HeartBeatRsp)
			pheart.SID = proto.String(req.GetSID())

			rsp.B = pheart
			fmt.Printf("RSP：\n%s\n", pro.Debug(rsp))

			err = n.SendData(conn, rsp)

			if err != nil {
				fmt.Printf("send hearbeat rsp, err = %s\n", err)
				return
			}
		}
	case pb.CMDHeartBeatRsp:
		{
		}
	case pb.CMDBizRsp:
		{
		}
	default:
		fmt.Printf("unknow command 0x%x\n", msg.H.Command)
	}
}

func clientRcv(conn *mynet.Connection) {
	n := conn.Net()
	data := conn.UserData.(*userData)
	w := data.w
	for {
		evt, err := n.PollEvent(1000 * 60)
		if err != nil {
			fmt.Printf("poll event error!")
			w.Done()
			return
		}
		conn := evt.Conn
		switch {
		case evt.EventType == mynet.EventConnectionError:
			{
				fmt.Printf("event error: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
			}
		case evt.EventType == mynet.EventConnectionClosed:
			{
				fmt.Printf("event close: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
			}
		case evt.EventType == mynet.EventNewConnectionData:
			{
				data := evt.Data.(*pb.PbProto)
				go handler(conn, data)
			}
		case evt.EventType == mynet.EventTimeout:
			{
				fmt.Printf("poll timeout now = %d\n", time.Now().Unix())
			}
		}
	}
}

func clientSend(conn *mynet.Connection) {
	n := conn.Net()
	data := conn.UserData.(*userData)
	w := data.w
	c := data.c
	pro := data.proto

	for {
		select {
		case msg := <-c:
			{
				switch msg {
				case "quit":
					{
						w.Done()
						return
					}
				case "heart":
					{
						m := &pb.PbProto{}
						m.H.Command = pb.CMDHeartBeatReq
						// beartbeat
						p, err := pro.GetMessage(m.H.Command)
						if err != nil {
							fmt.Printf("gen hearbeat failed, err = %s\n", err)
							continue
						}
						pheart := p.(*pb.HeartBeatReq)
						pheart.SID = proto.String(pb.SID(32))
						m.B = pheart

						fmt.Printf("SEND：\n%s\n", pro.Debug(m))
						err = n.SendData(conn, m)

						if err != nil {
							fmt.Printf("send hearbeat, err = %s\n", err)
							continue
						}
					}
				default:
					{
						m := &pb.PbProto{}
						m.H.Command = pb.CMDBizReq
						// beartbeat
						p, err := pro.GetMessage(m.H.Command)
						if err != nil {
							fmt.Printf("gen hearbeat failed, err = %s\n", err)
							continue
						}
						pbiz := p.(*pb.BizReq)
						pbiz.SID = proto.String(msg)
						m.B = pbiz

						fmt.Printf("SEND：\n%s\n", pro.Debug(m))
						err = n.SendData(conn, m)

						if err != nil {
							fmt.Printf("send biz, err = %s\n", err)
							continue
						}
					}
				}
			}
		}
	}
}

func input(conn *mynet.Connection) {
	reader := bufio.NewReader(os.Stdin)
	data := conn.UserData.(*userData)
	w := data.w
	c := data.c

	for {
		line, _, _ := reader.ReadLine()
		msg := string(line)
		c <- msg
		if msg == "quit" {
			w.Done()
			break
		}
	}
}

func main() {

	n := mynet.NewSimpleNet()

	c, e := n.Connect("127.0.0.1:3369")
	if e != nil {
		fmt.Printf("conn failed, err=%s\n", e)
		os.Exit(-1)
	}
	fmt.Printf("connect to 127.0.0.1:3369 success\n")

	data := &userData{
		w:     &sync.WaitGroup{},
		c:     make(chan string),
		proto: &pb.PbServerProto{},
	}
	c.Proto = data.proto
	c.UserData = data

	w := data.w

	w.Add(1)
	go clientSend(c)

	w.Add(1)
	go clientRcv(c)

	w.Add(1)
	go input(c)

	w.Wait()

	mynet.SimpleNetDestroy(n)
}
