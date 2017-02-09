package pb

import (
	"bytes"
	"encoding/binary"
	"fmt"

	mynet "github.com/buf1024/golib/net"
	"github.com/golang/protobuf/proto"
)

type Head struct {
	Command uint64
	Length  uint32
	Extral  uint64
}

type PbProto struct {
	H Head
	B proto.Message
}

type PbServerProto struct {
}

const (
	constHeadLen uint32 = 20
)

const (
	CMDHeartBeatReq uint64 = 0x00010001
	CMDHeartBeatRsp uint64 = 0x00010002
	CMDBizReq       uint64 = 0x00010003
	CMDBizRsp       uint64 = 0x00010004
)

var message map[uint64]proto.Message

func (p *PbServerProto) FilterAccept(conn *mynet.Connection) bool {
	return true
}
func (p *PbServerProto) HeadLen() uint32 {
	return constHeadLen
}
func (p *PbServerProto) BodyLen(head []byte) (interface{}, uint32, error) {
	if (uint32)(len(head)) != constHeadLen {
		return nil, 0, fmt.Errorf("head size not right")
	}
	h := &Head{}

	buf := head[:8]
	reader := bytes.NewReader(buf)
	err := binary.Read(reader, binary.BigEndian, &h.Command)
	if err != nil {
		return nil, 0, err
	}

	buf = head[8:12]
	reader = bytes.NewReader(buf)
	err = binary.Read(reader, binary.BigEndian, &h.Length)
	if err != nil {
		return nil, 0, err
	}

	buf = head[12:20]
	reader = bytes.NewReader(buf)
	err = binary.Read(reader, binary.BigEndian, &h.Extral)
	if err != nil {
		return nil, 0, err
	}
	return head, h.Length, nil
}
func (p *PbServerProto) Parse(head interface{}, body []byte) (interface{}, error) {
	m := &PbProto{
		H: *(head.(*Head)),
	}
	pb, err := getMessage(m.H.Command)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(body, pb)
	if err != nil {
		return nil, err
	}

	m.B = pb

	return m, nil
}
func (p *PbServerProto) Serialize(data interface{}) ([]byte, error) {
	m := data.(*PbProto)

	body, err := proto.Marshal(m.B)
	if err != nil {
		return nil, err
	}
	m.H.Length = (uint32)(len(body))

	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.BigEndian, m.H.Command); err != nil {
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, m.H.Length); err != nil {
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, m.H.Extral); err != nil {
		return nil, err
	}

	head := buf.Bytes()

	return append(head, body...), nil

}

func (p *PbServerProto) Debug(msg *PbProto) string {
	return fmt.Sprintf("command:0x%x\nlength :%d\nextral :%d\n==========\n%s",
		msg.H.Command, msg.H.Length, msg.H.Extral, proto.CompactTextString(msg.B))
}

func getMessage(command uint64) (proto.Message, error) {
	if m, ok := message[command]; ok {
		proto.Clone(m)
		return m, nil
	}
	return nil, fmt.Errorf("mommand %d not found", command)
}

func init() {
	message = make(map[uint64]proto.Message)

	message[CMDHeartBeatReq] = &HeartBeatReq{} //0x00010001 // 心跳请求
	message[CMDHeartBeatRsp] = &HeartBeatRsp{} //0x00010002

	message[CMDBizReq] = &BizReq{} //0x00010003
	message[CMDBizRsp] = &BizRsp{} //0x00010004
}
