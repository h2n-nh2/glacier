package glacier

/*
glacier
基于net.Conn的TCPServer,未来可以扩展到UDP,但是主要方向应是基于epoll的一个简单Server
Server 需要重写 IPacket,提供解包和处理过程
Server 设计需要继承使用
func (s *Server) ServeTcp
func (s *Server) Parse
func (s *Server) Handle
现在还有很多缺陷,慢慢整理
*/
import (
	"net"
	"sync"

	"github.com/rs/zerolog"
)

type Server struct {
	Addr       string
	Listen     net.Listener
	Dispatcher sync.Map
	Logger     *zerolog.Logger
}

// //////////////////////////// run & quit //////////////////////////////
func (s *Server) Quit() {
	s.Listen.Close()
	s.Dispatcher.Range(func(iden, h any) bool {
		h.(IHandler).Close()
		s.Dispatcher.Delete(iden)
		return true
	})
}

// ///////////////////////////////////////////////////////////////////////
func (s *Server) Log() *zerolog.Logger {
	return s.Logger
}

// func (s *Server) Parse(c io.Reader) (IPacket, error) {
// 	// no func ?
// 	s.Log().Error().Msgf("connection no real reader!!!!")
// 	return nil, nil
// }

func (s *Server) Send(pkt IPacket, iden interface{}) error {
	if h, ok := s.Dispatcher.Load(iden); ok {
		return h.(IHandler).Send(pkt)
	}
	return ERRCODE_NOTFOUND
}

func (s *Server) SendAll(pkt IPacket) error {
	s.Dispatcher.Range(func(iden, h any) bool {
		if h.(IHandler).IsIdentified() {
			h.(IHandler).Send(pkt)
		}
		// 忽略error
		return true
	})
	return nil
}

func (s *Server) Close(iden interface{}) {
	if h, ok := s.Dispatcher.Load(iden); ok {
		h.(IHandler).Close()
		s.Dispatcher.Delete(iden)
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////
// func (s *Server) Handle(pkt IPacket, h IHandler) {
// 	s.Log().Error().Msgf("server no real handler!!!!")
// }

func (s *Server) HandShake(iden interface{}, h IHandler) {
	h.SetIdentify(iden)
	old, ok := s.Dispatcher.Load(iden)
	if ok && old != h { // 可能遇到重发
		//
		s.Log().Warn().Msgf("new connect overtake old for (%+v) for [%+v]", iden, old.(IHandler))
		old.(IHandler).Close()
		s.Dispatcher.Delete(old)
	}
	s.Dispatcher.Store(iden, h)
	s.Log().Info().Msgf("identify connect %s to %+v", h.ConnectionInfo(), iden)
}

func (s *Server) OnConnected(h IHandler) {
	s.Log().Info().Msgf("get connect %s", h.ConnectionInfo())
}

func (s *Server) OnClosed(h IHandler) {
	s.Log().Info().Msgf("close connect %s", h.ConnectionInfo())
	iden := h.GetIdentify()
	if iden != nil {
		s.Dispatcher.Delete(iden)
	}
}

// //////////////////////////////////////////////////////////////////////////////////////////////////
