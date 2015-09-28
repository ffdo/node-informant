package alfred

import (
	"log"
	"net"
	"os"
)

type Alfred struct {
	socket    net.Conn
	localPath string
	Input     chan AlfredTLV
	Output    chan AlfredTLV
}

func NewAlfred(path string) (alfred Alfred, err error) {
	sType := "unix" // or "unixgram" or "unixpacket"
	laddr := net.UnixAddr{"/var/run/michael.run", sType}
	socket, err := net.DialUnix(sType, &laddr, /*can be nil*/
		&net.UnixAddr{path, sType})
	if err != nil {
		return
	}
	alfred = Alfred{socket: socket, localPath: "/var/run/michael.run"}
	return
}

func (a Alfred) Close() {
	log.Printf("Closing Socket")
	a.socket.Close()
	os.Remove(a.localPath)
}

func (a Alfred) Request(request AlfredTLV) (response chan AlfredTLV) {
	response = make(chan AlfredTLV)
	go func() {
		requestData, _ := request.Marshall()
		count, err := a.socket.Write(requestData)
		log.Printf("Written %d bytes to socket", count)
		if err != nil {
			log.Printf("Error while writing to socket: %v", err)
		}
		buf := make([]byte, 4)
		log.Printf("Trying to read TLV header from socket")
		count, err = a.socket.Read(buf)
		log.Printf("Read %d bytes: %v", count, buf)
		if err != nil || count != 4 {
			// TODO Throw error
			log.Printf("Error received wrong count of bytes %d or error %v", count, err)
		}
		tlv, err := UnmarshallTLVHeader(buf)
		if tlv.Type == STATUS_ERROR {
			log.Printf("Received error from alfred")
			close(response)
			return
		}
		log.Printf("Trying to read payload of %d bytes from socket", tlv.Length)
		buf = make([]byte, tlv.Length)
		count, err = a.socket.Read(buf)
		if err != nil || uint16(count) != tlv.Length {
			// TODO throw error
			log.Printf("Error received wrong count of bytes %s or error %v", count, err)
		}
		tlv.Data = buf
		response <- tlv
		close(response)
	}()
	return
}
