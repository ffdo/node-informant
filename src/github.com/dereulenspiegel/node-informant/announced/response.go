package announced

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

type Response struct {
	ClientAddr net.Addr
	Payload    []byte
}

type JsonAddr struct {
	IP   string
	Port int
	Zone string
}

type PrintableResponse struct {
	Addr  JsonAddr
	Bytes string
}

func (r Response) String() string {
	udpAddr := r.ClientAddr.(*net.UDPAddr)
	addr := JsonAddr{
		IP:   udpAddr.IP.String(),
		Port: udpAddr.Port,
		Zone: udpAddr.Zone,
	}
	var buffer bytes.Buffer
	for _, b := range r.Payload {
		buffer.WriteString(fmt.Sprintf("%d ", b))
	}
	printable := PrintableResponse{
		Addr:  addr,
		Bytes: buffer.String(),
	}
	data, _ := json.Marshal(printable)
	return string(data)
}
