package announced

import (
	"net"
)

type Response struct {
	ClientAddr net.Addr
	Payload    []byte
}
