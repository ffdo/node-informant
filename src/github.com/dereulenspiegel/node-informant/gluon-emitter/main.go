package main

import (
	"flag"
	"net"

	"github.com/dereulenspiegel/node-informant/gluon-emitter/data"
	"github.com/dereulenspiegel/node-informant/utils"

	log "github.com/Sirupsen/logrus"

	"regexp"
)

var (
	interfaceFlag = flag.String("interface", "", "Specify the interface to listen on")
	portFlag      = flag.Int("port", 1101, "Specify the port to listen on")
	aliasFilePath = flag.String("alias", "", "Alias file to load")

	requestRegexp = regexp.MustCompile(`^GET\s([\w]+)$`)
)

func main() {
	flag.Parse()
	err := data.LoadYamlFile(*aliasFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  *aliasFilePath,
			"error": err,
		}).Fatal("Can't load alias file")
	}
	localIp, err := utils.GetIPFromInterface(*interfaceFlag)
	if err != nil {
		log.WithFields(log.Fields{
			"interface": *interfaceFlag,
			"error":     err,
		}).Fatal("Can't determine IP address for interface")
	}

	localAddr := &net.UDPAddr{
		IP:   *localIp,
		Port: *portFlag,
	}

	conn, err := net.ListenUDP("udp6", localAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"interface":    *interfaceFlag,
			"port":         *portFlag,
			"localAddress": *localAddr,
			"error":        err,
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
		} else {
			if requestRegexp.Match(buf[0:n]) {
				finds := requestRegexp.FindAllSubmatch(buf[:n], -1)
				section := string(finds[0][1])
				out, err := data.GetMarshalledAndCompressedSection(section)
				if err != nil {
					n, err := conn.WriteToUDP(out, addr)
					if err != nil {
						log.WithFields(log.Fields{
							"remoteAddress": *addr,
							"error":         err,
						}).Error("Can't write requested data to remote")
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
