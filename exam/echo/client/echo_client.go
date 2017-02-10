package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"bufio"

	mynet "github.com/buf1024/golib/net"
)

type userData struct {
	w *sync.WaitGroup
	c chan string
}

func clientRcv(conn *mynet.Connection) {
	n := conn.Net()
	data := conn.UserData.(*userData)
	w := data.w
	c := data.c
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
				close(c)
				w.Done()
				return
			}
		case evt.EventType == mynet.EventConnectionClosed:
			{
				fmt.Printf("event close: local = %s, remote = %s\n",
					conn.LocalAddress(), conn.RemoteAddress())
				close(c)
				w.Done()
				return
			}
		case evt.EventType == mynet.EventNewConnectionData:
			{
				data := evt.Data.([]byte)
				fmt.Printf("client receve: %s\n", (string)(data))
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

	for {
		select {
		case msg, ok := <-c:
			{
				if !ok {
					w.Done()
					return
				}
				if msg == "quit" {
					n.CloseConn(conn)
					w.Done()
					return
				}
				err := n.SendData(conn, ([]byte)(msg))
				if err != nil {
					fmt.Printf("send data error, err = %s\n", err)
					break
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
		if conn.Status() != mynet.StatusConnected {
			w.Done()
			break
		}
		c <- msg
		if msg == "quit" {
			w.Done()
			break
		}
	}
}

func main() {

	n := mynet.NewSimpleNet()

	c, e := n.Connect("127.0.0.1:3369", nil)
	if e != nil {
		fmt.Printf("conn failed, err=%s\n", e)
		os.Exit(-1)
	}
	fmt.Printf("connect to 127.0.0.1:3369 success\n")

	data := &userData{
		w: &sync.WaitGroup{},
		c: make(chan string),
	}
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

	fmt.Printf("client Done\n")
}
