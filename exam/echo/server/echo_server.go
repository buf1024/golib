package main

import (
	"fmt"
	"os"

	"time"

	mynet "github.com/buf1024/golib/net"
)

func server(listen *mynet.Listener) {
	fmt.Printf("server listenning %s\n", listen.LocalAddress())
	n := listen.Net()
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
			}
		case evt.EventType == mynet.EventNewData:
			{
				data := evt.Data.([]byte)
				fmt.Printf("%s", (string)(data))
				err = n.SendData(evt.Conn, evt.Data)
				if err != nil {
					fmt.Printf("send data error, err = %s\n", err)
				}
			}
		case evt.EventType == mynet.EventNewConnected:
			{
				fmt.Printf("client conneced: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
			}
		case evt.EventType == mynet.EventTimeout:
			{
				fmt.Printf("poll timeout now = %d\n", time.Now().Unix())
			}
		}
	}
}

func main() {

	n := mynet.NewSimpleNet()

	s, e := n.Listen("127.0.0.1:3369")
	if e != nil {
		fmt.Printf("listen failed, err=%s\n", e)
		os.Exit(-1)
	}

	server(s)

	mynet.SimpleNetDestroy(n)

}
