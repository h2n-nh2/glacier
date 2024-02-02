package glacier

import (
	"io"

	"github.com/rs/zerolog"
)

// 当前的fzacket处理比较特别,解决不具有通用型
type IPacket interface {
	// Read(buf io.Reader) (int64, error)
	Write(buf io.Writer) (int64, error)

	// 丧失通用型
	GetLen() uint32 // Gets the length of the message data segment(获取消息数据段长度)
	GetType() uint8
	GetData() []byte // skip header
	// GetData() []byte // Gets the content of the message(获取消息内容)
	// GetRawData() []byte // Gets the raw data of the message(获取原始数据)
	// SetData([]byte)     // Sets the content of the message(设计消息内容)
	// SetDataLen(uint32)  // Sets the length of the message data segment(设置消息数据段长度)
}

type IHandler interface {
	Process()
	Close()
	Send(IPacket) error
	IsIdentified() bool
	SetIdentify(iden interface{})
	GetIdentify() interface{}
	ConnectionInfo() string
	SetAttachment(a interface{})
	GetAttachment() interface{}
}

type IServer interface {
	ServeTcp(address string) (err error)
	Quit()
	//CreateHandler(c net.Conn) IHandler
	Parse(buf io.Reader) (IPacket, error)

	Send(pkt IPacket, iden interface{}) error
	SendAll(pkt IPacket) error
	Close(iden interface{})

	Handle(pkt IPacket, h IHandler) error
	HandShake(iden interface{}, h IHandler)

	OnConnected(h IHandler)
	OnClosed(h IHandler)
	Log() *zerolog.Logger
}
