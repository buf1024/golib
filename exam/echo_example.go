package main

import (
	"fmt"
	"os"
	"sync"

	mynet "github.com/buf1024/golib/net"
)

type echoProto struct {
}

func (p *echoProto) HeadLen() int64 {
	return 0
}
func (p *echoProto) BodyLen(head []byte) int64 {
	return 0
}

func (p *echoProto) Parse(head []byte, body []byte) (interface{}, error) {
	return nil, nil
}

func (p *echoProto) Serialize(data interface{}) ([]byte, error) {
	return nil, nil
}

func server(conn *mynet.Connection) {

}

func client(conn *mynet.Connection) {

}

func input() {

}

func main() {
	w := &sync.WaitGroup{}

	n := mynet.NewSimpleNet(&echoProto{}, w)

	s, e := n.Listen("127.0.0.1:3369")
	if e != nil {
		fmt.Printf("listen failed, err=%s\n", e)
		os.Exit(-1)
	}

	w.Add(1)
	go server(s)

	c, e := n.Connect("127.0.0.0:3369")
	if e != nil {
		fmt.Printf("conn failed, err=%s\n", e)
		os.Exit(-1)
	}
	c.UserData = make(chan string)

	w.Add(1)
	go client(c)

	w.Add(1)
	go input()

	w.Wait()

}
