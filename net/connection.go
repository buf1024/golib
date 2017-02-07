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
	Net *SimpleNet

	ID         int64
	Type       int64
	Status     int64
	Conn       net.Conn
	LocalAddr  string
	RemoteAddr string
	Proto      IProto // 为了实现多种proto

	UserData interface{}
}

type SimpleNet struct {
	proto  IProto
	conns  map[int64]*Connection
	events chan *ConnEvent

	UserData interface{}
}

type IProto interface {
	HeadLen() int64
	BodyLen(head []byte) int64
	Parse(head []byte, body []byte) (interface{}, error)
	Serialize(data interface{}) ([]byte, error)
}

func NewSimpleNet(proto IProto, userData interface{}) *SimpleNet {
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

func (c *SimpleNet) SendData(conn *Connection, data interface{}) error {
	return nil
}
