package announced

import (
	"fmt"
	"io"
	"net"

	log "github.com/Sirupsen/logrus"
)

// MultiCastGroup is the default multicast group used by announced
const MultiCastGroup string = "ff02:0:0:0:0:0:2:1001"

// Port is the default udp port used by announced
const Port int = 1001

// Proto specifies that announced will only work with UDP on IPv6
const Proto string = "udp6"

// MaxDataGramSize is more a less a guessed value of the maximum receivable size
const MaxDataGramSize int = 8192

var announcedAddr = &net.UDPAddr{IP: net.ParseIP(MultiCastGroup), Port: Port}

// AnnouncedPacketReceiver abstracts the receiption of packets on the network side
// away so we can mock this easily in tests.
type AnnouncedPacketReceiver interface {
	io.Closer
	// Receive registers a callback method called every time packet is delivered
	// Normally this method jusz enqueues the Repsonse in a channel for further processing.
	Receive(rFunc func(Response))
	Query(queryString string)
	QueryUnicast(addr *net.UDPAddr, queryString string)
}

// Query represents who and what to query. If TargetAddr is null the default
// multicast address will be used to query all nodes in this multicast group.
type Query struct {
	TargetAddr  *net.UDPAddr
	QueryString string
}

// Requester is responsible for sending out queries and receiving the responses.
// The requester does not process the Responses in any way.
type Requester struct {
	unicastConn net.PacketConn
	queryChan   chan Query
	ReceiveChan chan Response
}

// getIPFromInterface tries to determine the link local IPv6 unicast address of
// an interface named by the given string. Returns an error if the interface is
// not found, or the interface has not a link local IPv6 unicast address (i.e.
// because IPv6 is not configured for this interface).
func getIPFromInterface(ifaceName string) (*net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, err
	}

	addresses, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addresses {
		ip, ok := addr.(*net.IPNet)
		if ok {
			if ip.IP.To4() == nil && ip.IP.IsLinkLocalUnicast() {
				return &ip.IP, nil
			}
		}
	}
	return nil, fmt.Errorf("No valid IPv6 address found on interface %s", ifaceName)
}

// NewRequester creates a new Requester using the interface named by interfaceName
// and listening on the port specified for responses.
func NewRequester(ifaceName string, port int) (r *Requester, err error) {
	var lIP *net.IP = &net.IPv6zero
	r = &Requester{}
	if ifaceName != "" {
		lIP, err = getIPFromInterface(ifaceName)
		if err != nil {
			return
		}
	} else {
		err = fmt.Errorf("No interface specified")
		return
	}
	r.unicastConn, err = net.ListenPacket(Proto, fmt.Sprintf("[%s%%%s]:%d", lIP.String(), ifaceName, port))
	if err != nil {
		return
	}
	r.queryChan = make(chan Query)
	r.ReceiveChan = make(chan Response, 100)
	go r.writeLoop()
	go r.readLoop()
	return
}

// writeLoop waits for Queries on a channel and writes the immediately to the
// socket.
func (r *Requester) writeLoop() {
	for query := range r.queryChan {
		queryString := query.QueryString
		targetAddr := query.TargetAddr
		if targetAddr == nil {
			targetAddr = announcedAddr
		}
		buf := []byte(queryString)
		count, err := r.unicastConn.WriteTo(buf, targetAddr)
		if count < len(buf) {
			log.Printf("Written less bytes (%d) than expected (%d)", count, len(buf))
			log.WithFields(log.Fields{
				"bytesWritten":  count,
				"bytesExpected": len(buf),
			}).Error("Failed to write all bytes to unicast address")
		}
		if err != nil {
			log.Printf("Error while writing to MulticastGroup: %v", err)
			log.WithFields(log.Fields{
				"multicastGroup": announcedAddr,
				"error":          err,
			}).Error("Error writing to multicast group")
		}
	}
}

// readLoop reads UDP packets from the socket and puts these Respones on a channel
func (r *Requester) readLoop() {
	buf := make([]byte, MaxDataGramSize)
	for {
		count, raddr, err := r.unicastConn.ReadFrom(buf)
		if err != nil {
			log.Printf("Error reading from MulticastGroup: %v", err)
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error reading from udp socket, closing")
			break
		}
		payload := make([]byte, count)
		copy(payload, buf)
		response := Response{
			ClientAddr: raddr,
			Payload:    payload,
		}
		r.ReceiveChan <- response
	}
}

// Close closes the Requester instance and frees all allocated resources
func (r *Requester) Close() error {
	r.unicastConn.Close()
	close(r.ReceiveChan)
	close(r.queryChan)
	return nil
}

// QueryUnicast sends an UDP query to a host directly via unicast. The IPv6 address
// and the port where announced listens on the remote node need to be known.
func (r *Requester) QueryUnicast(addr *net.UDPAddr, queryString string) {
	query := Query{QueryString: queryString, TargetAddr: addr}
	r.queryChan <- query
}

// Query multicasts the specified query to the default announced multicast group
// on the default port.
func (r *Requester) Query(queryString string) {
	query := Query{QueryString: queryString}
	r.queryChan <- query
}

// Receive is an implementation of the AnnouncedPacketReceiver interface
func (r *Requester) Receive(rFunc func(Response)) {
	for response := range r.ReceiveChan {
		rFunc(response)
	}
}
