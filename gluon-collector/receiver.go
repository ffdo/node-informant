package main

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	cfg "github.com/olebedev/config"

	"github.com/dereulenspiegel/node-informant/announced"
)

type MultiReceiver struct {
	packetChan    chan announced.Response
	childReceiver []announced.AnnouncedPacketReceiver
}

func NewMultiReceiver(receivers ...announced.AnnouncedPacketReceiver) *MultiReceiver {
	mr := &MultiReceiver{make(chan announced.Response, 100), make([]announced.AnnouncedPacketReceiver, 0, 2)}
	mr.childReceiver = append(mr.childReceiver, receivers...)
	for _, receiver := range receivers {
		go mr.singleReceive(receiver)
	}
	return mr
}

func (m *MultiReceiver) Query(queryString string) {
	for _, receiver := range m.childReceiver {
		receiver.Query(queryString)
	}
}

func (m *MultiReceiver) QueryUnicast(addr *net.UDPAddr, queryString string) {
	for _, receiver := range m.childReceiver {
		receiver.QueryUnicast(addr, queryString)
	}
}

func (m *MultiReceiver) singleReceive(receiver announced.AnnouncedPacketReceiver) {
	receiver.Receive(func(response announced.Response) {
		m.packetChan <- response
	})
}

func (m *MultiReceiver) Receive(rFunc func(announced.Response)) {
	for packet := range m.packetChan {
		rFunc(packet)
	}
}

func (m *MultiReceiver) Close() error {
	for _, receiver := range m.childReceiver {
		receiver.Close()
	}
	return nil
}

func buildReceiver() announced.AnnouncedPacketReceiver {
	receiverConfigList, err := conf.Global.List("receiver")
	if err != nil {
		log.Fatalf("Receiver don't seem to be configured: %v", err)
	}
	receiverCount := len(receiverConfigList)
	receiverSlice := make([]announced.AnnouncedPacketReceiver, 0, receiverCount)
	for i := 0; i < receiverCount; i++ {
		receiverConfig, err := conf.Global.Get(fmt.Sprintf("receiver.%d", i))
		if err != nil {
			log.Fatalf("Error retrieving config for %dth receiver: %v", i, err)
		}
		receiver := receiverFactory(receiverConfig)
		receiverSlice = append(receiverSlice, receiver)
	}

	return NewMultiReceiver(receiverSlice...)
}

func receiverFactory(receiverConfig *cfg.Config) announced.AnnouncedPacketReceiver {
	receiverType, err := receiverConfig.String("type")
	if err != nil {
		log.Fatalf("Can't retrieve type of receiver: %v", err)
	}

	switch receiverType {
	case "announced":
		return buildAnnouncedReceiver(receiverConfig)
	default:
		log.Fatalf("Unknown receiver type %s", receiverType)
		return nil
	}
}

func buildAnnouncedReceiver(announcedConfig *cfg.Config) announced.AnnouncedPacketReceiver {
	iface, err := announcedConfig.String("interface")
	if err != nil {
		log.Fatalf("Can't determine interface for announced receiver")
	}

	port, err := announcedConfig.Int("port")
	if err != nil {
		log.Fatalf("Can't determine port for announced receiver")
	}
	requester, err := announced.NewRequester(iface, port)
	if err != nil {
		log.Fatalf("Error creating requester: %v", err)
	}
	return requester
}
