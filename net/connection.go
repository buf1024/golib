package net

import (
	"net"
)

const (
	EventNone = iota
	EventNewConnected
	EventDisconnected
	EventError
	EventNewData
)

const (
	TypeNone = iota
	TypeServer
	TypeClient
)

const (
	StatusNone = iota
	StatusConnected
	StatusBroken
)

type ConnEvent struct {
	EventType int64
	Conn      *Connection
	Data      interface{}
}

type Connection struct {
	mynet *SimpleNet

	ID         int64
	Type       int64
	Status     int64
	Conn       net.Conn
	LocalAddr  string
	RemoteAddr string

	UserData interface{}
}

type SimpleNet struct {
	proto  IProto
	conns  map[int64]*Connection
	events chan *ConnEvent
}

type IProto interface {
	HeadLen() int64
	BodyLen(head []byte) int64
	Parse(head []byte, body []byte) (interface{}, error)
	Serialize(head []byte, body []byte) error
}

func NewSimpleNet(proto IProto) *SimpleNet {
	return &SimpleNet{}
}

func (c *SimpleNet) Listen(host string) (*Connection, error) {
	return nil, nil
}
func (c *SimpleNet) Connect(host string) (*Connection, error) {
	return nil, nil
}

func (c *SimpleNet) PollEvent() (*ConnEvent, error) {
	return nil, nil
}

func (c *SimpleNet) SendData(conn *Connection, data []byte) error {
	return nil
}
