package glacier

import (
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	CONN_Base_TimeOutMS       = 4500
	Handle_Seq          int32 = 0
)

type GlaCierError int16

const (
	ERRCODE_SUCCESS        GlaCierError = 0
	ERRCODE_UNKNOWN        GlaCierError = -1
	ERRCODE_DISCONNECT     GlaCierError = -10
	ERRCODE_TIMEOUT        GlaCierError = -11
	ERRCODE_READFAIL       GlaCierError = -12
	ERRCODE_WRITEFAIL      GlaCierError = -13
	ERRCODE_QUEUEFULL      GlaCierError = -14
	ERRCODE_NOTFOUND       GlaCierError = -20
	ERRCODE_CALLBACK_UNREG GlaCierError = -21
	ERRCODE_PROTO          GlaCierError = -22
)

// error define 以后要详细设定,现在没有用好
func (r GlaCierError) Error() string {
	switch r {
	case ERRCODE_SUCCESS:
		return "success"
	case ERRCODE_UNKNOWN: // not use yet
		return "unknown"
	case ERRCODE_DISCONNECT: // not use yet
		return "disconnect"
	case ERRCODE_TIMEOUT: // not use yet
		return "timeout"
	case ERRCODE_READFAIL: // not use yet
		return "read fail"
	case ERRCODE_WRITEFAIL: // not use yet
		return "write fail"
	case ERRCODE_QUEUEFULL:
		return "queue full"
	case ERRCODE_NOTFOUND:
		return "no connect found"
	case ERRCODE_CALLBACK_UNREG:
		return "command not found"
	case ERRCODE_PROTO:
		return "noncompliance with protocol"
	default:
		return strconv.Itoa(int(r))
	}
}

// base go routings for server
type Handler struct {
	ID        int32
	Iden      interface{}
	Server    IServer
	Conn      net.Conn
	atta      interface{}
	Queue     chan IPacket
	TimeOutMs int
}

func NewHandler(c net.Conn, s IServer, HStimeOutms int) IHandler {
	a := new(Handler)
	a.Iden = nil
	a.Server = s
	a.Conn = c
	a.Queue = make(chan IPacket, 10)
	if HStimeOutms > 0 {
		a.TimeOutMs = HStimeOutms
	} else {
		a.TimeOutMs = CONN_Base_TimeOutMS
	}

	a.atta = nil
	a.ID = atomic.AddInt32(&Handle_Seq, 1)
	return a
}

func (h *Handler) Process() {
	go h.writing(h.Server)
	h.reading(h.Server)
}

func (h *Handler) Close() {
	close(h.Queue)
}

func (h *Handler) Send(pkt IPacket) error {
	select {
	case h.Queue <- pkt:
		return ERRCODE_SUCCESS
	default:
		return ERRCODE_QUEUEFULL
	}
}

func (h *Handler) IsIdentified() bool {
	return h.Iden != nil
}

func (h *Handler) SetIdentify(iden interface{}) {
	h.Iden = iden
}

func (h *Handler) GetIdentify() interface{} {
	return h.Iden
}

func (h *Handler) ConnectionInfo() string {
	return h.Conn.RemoteAddr().String()
}

func (h *Handler) reading(s IServer) {
	s.OnConnected(h)
	for {
		if h.Iden == nil { // 初始包有超时
			h.Conn.SetDeadline(time.Now().Add(time.Duration(h.TimeOutMs) * time.Millisecond))
		} else {
			h.Conn.SetDeadline(time.Time{})
		}
		//
		pkt, err := s.Parse(h.Conn)
		if err != nil {
			s.Log().Error().Msgf("handler [%+v](%d) error recv packet from net conn: %s", h.Iden, h.ID, err)
			break
		}
		//防止数据并行处理，导致错乱
		//s.Log().Info().Msgf("handler [%+v](%d) receive (%d) bytes", h.Iden, h.ID, pkt.GetLen())
		if err := h.Server.Handle(pkt, h); err != nil {
			//
			break
		}
	}
	close(h.Queue)
}

func (h *Handler) writing(s IServer) {
	for pkt := range h.Queue {
		if pkt == nil {
			s.Log().Info().Msgf("handler [%+v](%d) queue channel closed", h.Iden, h.ID)
			break
		}
		// 注意这里可能不够通用,如果发包过大,数据会写不完EAGAIN
		if _, err := pkt.Write(h.Conn); err != nil {
			s.Log().Error().Msgf("handler [%+v](%d) error send packet into net conn: %s", h.Iden, h.ID, err)
			break
		}
	}
	h.Conn.Close()
	s.OnClosed(h)
}

func (h *Handler) SetAttachment(a interface{}) {
	h.atta = a
}

func (h *Handler) GetAttachment() interface{} {
	return h.atta
}
