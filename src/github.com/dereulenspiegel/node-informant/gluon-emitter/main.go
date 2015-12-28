package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector/hostname"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector/uptime"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/data"
	"golang.org/x/net/ipv6"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"regexp"
)

var (
	interfaceFlag    = flag.String("interface", "", "Specify the interface to listen on")
	portFlag         = flag.Int("port", 1101, "Specify the port to listen on")
	aliasFilePath    = flag.String("alias", "", "Alias file to load")
	groupAddressFlag = flag.String("group", "ff02:0:0:0:0:0:2:1001", "Address of multicast group to join")
	logLevelFlag     = flag.String("loglevel", "info", "Set the log level")
	configFileFlag   = flag.String("config", "", "Path to a configuration yaml file")

	requestRegexp = regexp.MustCompile(`^GET\s([\w]+)$`)
)

func init() {
	hostname.Init()
	uptime.Init()
}

func parseConfig(filePath string) map[string]interface{} {
	configMap := make(map[string]interface{})
	configBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"configFilePath": filePath,
			"error":          err,
		}).Fatal("Can't read configuration")
	}

	err = yaml.Unmarshal(configBytes, &configMap)
	if err != nil {
		log.WithFields(log.Fields{
			"error":          err,
			"configFilePath": filePath,
		}).Fatal("Can't unmarshal configuration file")
	}
	normalizedMap, err := data.NormalizeMap(configMap)
	if err != nil {
		log.Fatalf("Can't normalize config data: %v", err)
	}
	return normalizedMap.(map[string]interface{})
}

func joinMulticastGroup(interfaceName, groupAddr string, port int) (*ipv6.PacketConn, error) {
	log.WithFields(log.Fields{
		"interface": interfaceName,
		"group":     groupAddr,
		"port":      port,
	}).Info("Joining multicast group")
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	group := net.ParseIP(groupAddr)
	c, err := net.ListenPacket("udp6", fmt.Sprintf("[::]:%d", port))
	if err != nil {
		return nil, err
	}
	p := ipv6.NewPacketConn(c)

	if err := p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
		return nil, err
	}
	if err := p.SetControlMessage(ipv6.FlagDst, true); err != nil {
		return nil, err
	}
	return p, nil
}

func main() {
	flag.Parse()
	config := parseConfig(*configFileFlag)
	logLevel, err := log.ParseLevel(*logLevelFlag)
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Info("Loading alias file")
	err = data.LoadAliases(*aliasFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  *aliasFilePath,
			"error": err,
		}).Fatal("Can't load alias file")
	}
	collector.InitCollection(config)

	conn, err := joinMulticastGroup(*interfaceFlag, *groupAddressFlag, *portFlag)
	if err != nil {
		log.WithFields(log.Fields{
			"interface":    *interfaceFlag,
			"port":         *portFlag,
			"groupAddress": *groupAddressFlag,
			"error":        err,
		}).Fatal("Can't open server socket")
	}
	log.Info("Joined multicast group.")

	buf := make([]byte, 1500)
	for {
		log.Debug("Waiting for packets...")
		n, _, src, err := conn.ReadFrom(buf)
		if err != nil {
			log.WithFields(log.Fields{
				"receivedByteCount": n,
				"error":             err,
				"remoteAddress":     src,
			}).Error("Error reading from UDP connection")
			return
		} else {
			log.WithFields(log.Fields{
				"data":          buf[:n],
				"byteCount":     n,
				"remoteAddress": src,
			}).Debug("Received data on multicast group")

			if requestRegexp.Match(buf[0:n]) {
				finds := requestRegexp.FindAllSubmatch(buf[:n], -1)
				section := string(finds[0][1])
				log.WithFields(log.Fields{
					"section": section,
				}).Debug("Received valid request for section")
				response, err := data.GetMarshalledAndCompressedSection(section)
				if err == nil {
					log.Debugf("Found data to respond, responding %d bytes", len(response))
					n, err := conn.WriteTo(response, nil, src)
					if err != nil {
						log.WithFields(log.Fields{
							"remoteAddress": src,
							"error":         err,
						}).Error("Can't write requested data to remote")
						return
					}

					if n != len(response) {
						log.WithFields(log.Fields{
							"dataLength":    len(response),
							"bytesWritten":  n,
							"remoteAddress": src,
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
	log.Info("gluon-emitter is exiting")
}
