package announced

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
)

// Response represents the raw response received from announced.
// Note that the Playoad may be modified by different parts of the
// receive pipe, i.e. deflating the content in case it is compressed.
// So take care of the order of the pipes in the receibe pipeline.
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

// Prints a received raw response as a string. The Payload will also
// be represented as a string. If it is not a string, this will be rather
// useless.
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
