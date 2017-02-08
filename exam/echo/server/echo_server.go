package main

import (
	"fmt"
	"os"
	"sync"

	mynet "github.com/buf1024/golib/net"
)

func server(listen *mynet.Listener) {
	n := listen.Net
	w := (n.UserData).(*sync.WaitGroup)
	for {
		evt, err := n.PollEvent()
		if err != nil {
			fmt.Printf("poll event error!")
			w.Done()
			return
		}
		if evt.EventType == mynet.EventError {
			fmt.Printf("event error: local = %s, remote = %s\n",
				evt.Conn.LocalAddr, evt.Conn.RemoteAddr)
		}
		if evt.EventType == mynet.EventNewData {
			fmt.Printf("server receve: %s\n", evt.Data.(string))
			err = n.SendData(evt.Conn, evt.Data)
			if err != nil {
				fmt.Printf("send data error, err = %s\n", err)
			}
		}
	}
}

func main() {
	w := &sync.WaitGroup{}

	n := mynet.NewSimpleNet(nil, w)

	s, e := n.Listen("127.0.0.1:3369")
	if e != nil {
		fmt.Printf("listen failed, err=%s\n", e)
		os.Exit(-1)
	}

	w.Add(1)
	go server(s)

	w.Wait()

}
