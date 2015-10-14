package annunced

import (
	"net"
	"runtime"
)

type Server struct {
	conn        *net.UDPConn
	ReceiveChan chan Response
}

func NewServer(ifaceName string) (*Server, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, err
	}
	con, err := net.ListenMulticastUDP(Proto, iface, announcedAddr)
	if err != nil {
		return nil, err
	}
	serv := &Server{conn: con, ReceiveChan: make(chan Response, 5)}
	return serv, nil
}

func (s *Server) readLoop() {
	var socketIsOpen = true
	var buf []byte = make([]byte, MaxDataGramSize)
	for socketIsOpen {
		count, raddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Error reading from MulticastGroup: %v", err)
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error reading from udp socket, closing")
			socketIsOpen = false
			continue
		}
		payload := make([]byte, count)
		copy(payload, buf)
		response := Response{
			ClientAddr: raddr,
			Payload:    payload,
		}
		s.ReceiveChan <- response
		runtime.Gosched()
	}
}
