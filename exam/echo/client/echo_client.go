package main

import (
	"fmt"
	"os"
	"sync"

	"bufio"

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

func client(conn *mynet.Connection) {
	n := conn.Net
	msgChan := conn.UserData.(chan string)
	for {
		select {
		case msg := <-msgChan:
			{
				if msg == "quit" {
					close(msgChan)
					return
				}
				err := n.SendData(conn, ([]byte)(msg))
				if err != nil {
					fmt.Printf("send data error, err = %s\n", err)
				}
			}
		}
	}
}

func input(conn *mynet.Connection) {
	msgChan := conn.UserData.(chan string)
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _, _ := reader.ReadLine()
		msg := string(line)
		msgChan <- msg
		if msg == "quit" {
			break
		}
	}
}

func main() {
	w := &sync.WaitGroup{}

	n := mynet.NewSimpleNet(nil, w)

	c, e := n.Connect("127.0.0.0:3369")
	if e != nil {
		fmt.Printf("conn failed, err=%s\n", e)
		os.Exit(-1)
	}
	c.UserData = make(chan string)

	w.Add(1)
	go client(c)

	w.Add(1)
	go input(c)

	w.Wait()

}
