package announced

import (
	"fmt"
	"net"
	"runtime"

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

type AnnouncedPacketReceiver interface {
	Receive(rFunc func(Response))
}

type Query struct {
	TargetAddr  *net.UDPAddr
	QueryString string
}

type Requester struct {
	unicastConn net.PacketConn
	queryChan   chan Query
	ReceiveChan chan Response
}

func getIPFromInterface(ifaceName string) (*net.IP, error) {
	iface, _ := net.InterfaceByName(ifaceName)

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
	return nil, fmt.Errorf("No valid IPv6 address found on interface")
}

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
		//runtime.Gosched()
	}
}

func (r *Requester) readLoop() {
	var socketIsOpen = true
	var buf []byte = make([]byte, MaxDataGramSize)
	for socketIsOpen {
		count, raddr, err := r.unicastConn.ReadFrom(buf)
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
		r.ReceiveChan <- response
		runtime.Gosched()
	}
}

func (r *Requester) Close() {
	r.unicastConn.Close()
	close(r.ReceiveChan)
	close(r.queryChan)
}

func (r *Requester) QueryUnicast(addr *net.UDPAddr, queryString string) {
	query := Query{QueryString: queryString, TargetAddr: addr}
	r.queryChan <- query
}

func (r *Requester) Query(queryString string) {
	query := Query{QueryString: queryString}
	r.queryChan <- query
}

func (r *Requester) Receive(rFunc func(Response)) {
	for response := range r.ReceiveChan {
		rFunc(response)
	}
}
