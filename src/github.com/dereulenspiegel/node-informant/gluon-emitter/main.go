package main

import (
	"flag"
	"net"

	"github.com/dereulenspiegel/node-informant/gluon-emitter/data"

	log "github.com/Sirupsen/logrus"

	"regexp"
)

var (
	interfaceFlag   = flag.String("interface", "", "Specify the interface to listen on")
	portFlag        = flag.Int("port", 1101, "Specify the port to listen on")
	aliasFilePath   = flag.String("alias", "", "Alias file to load")
	goupAddressFlag = flag.String("group", "ff02:0:0:0:0:0:2:1001", "Address of multicast group to join")

	requestRegexp = regexp.MustCompile(`^GET\s([\w]+)$`)
)

func main() {
	flag.Parse()
	log.Info("Loading alias file")
	err := data.LoadAliases(*aliasFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  *aliasFilePath,
			"error": err,
		}).Fatal("Can't load alias file")
	}

	iface, err := net.InterfaceByName(*interfaceFlag)
	if err != nil {
		log.WithFields(log.Fields{
			"interfaceName": *interfaceFlag,
			"error":         err,
		}).Fatal("Can't find interface to listen on")
	}

	multicastGroupIp := net.ParseIP(*goupAddressFlag)
	groupAddr := &net.UDPAddr{
		IP:   multicastGroupIp,
		Port: *portFlag,
	}

	log.WithFields(log.Fields{
		"multicastAddress": *goupAddressFlag,
		"port":             *portFlag,
	}).Info("Joining multicast group")
	conn, err := net.ListenMulticastUDP("udp6", iface, groupAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"interface": *interfaceFlag,
			"port":      *portFlag,
			"error":     err,
		}).Fatal("Can't open server socket")
	}

	buf := make([]byte, 0, 1024)
	for true {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.WithFields(log.Fields{
				"receivedByteCount": n,
				"error":             err,
				"remoteAddress":     *addr,
			}).Error("Error reading from UDP connection")
			return
		} else {
			log.WithFields(log.Fields{
				"data":          buf[:n],
				"remoteAddress": *addr,
			}).Debug("Received data on multicast group")

			if requestRegexp.Match(buf[0:n]) {
				finds := requestRegexp.FindAllSubmatch(buf[:n], -1)
				section := string(finds[0][1])
				log.WithFields(log.Fields{
					"section": section,
				}).Debug("Received valid request for section")
				out, err := data.GetMarshalledAndCompressedSection(section)
				if err != nil {
					n, err := conn.WriteToUDP(out, addr)
					if err != nil {
						log.WithFields(log.Fields{
							"remoteAddress": *addr,
							"error":         err,
						}).Error("Can't write requested data to remote")
						return
					}

					if n != len(out) {
						log.WithFields(log.Fields{
							"dataLength":    len(out),
							"bytesWritten":  n,
							"remoteAddress": *addr,
						}).Error("Written less bytes than expected to remote")
					}
				} else {
					log.WithFields(log.Fields{
						"section": section,
						"error":   err,
					}).Error("Failed to retrieve requested data")
				}
			}
		}
	}
}
