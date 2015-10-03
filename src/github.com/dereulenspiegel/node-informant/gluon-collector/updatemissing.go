package main

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type MissingUpdater struct {
	Requester *announced.Requester
	Store     data.Nodeinfostore
}

func (m *MissingUpdater) UpdateMissingNeighbours() {
	log.Print("Updating missing neighbour infos")

	for _, nodeinfo := range m.Store.GetNodeInfos() {
		_, err := m.Store.GetNodeNeighbours(nodeinfo.NodeId)
		if err != nil {
			log.Debugf("Updating missing neighbour information for Node %s", nodeinfo.NodeId)
			for _, addressString := range nodeinfo.Network.Addresses {
				ip := net.ParseIP(addressString)
				addr := &net.UDPAddr{
					IP:   ip,
					Port: 1001,
				}
				log.Debugf("Querying IP %s for missing neighbours", addressString)
				m.Requester.QueryUnicast(addr, "GET neighbours")
			}
		}
	}
}

func (m *MissingUpdater) UpdateMissingStatistics() {
	log.Print("Updating missing neighbour infos")

	for _, nodeinfo := range m.Store.GetNodeInfos() {
		_, err := m.Store.GetStatistics(nodeinfo.NodeId)
		if err != nil {
			log.Debugf("Updating missing statistics information for Node %s", nodeinfo.NodeId)
			for _, addressString := range nodeinfo.Network.Addresses {
				ip := net.ParseIP(addressString)
				addr := &net.UDPAddr{
					IP:   ip,
					Port: 1001,
				}
				log.Debugf("Querying IP %s for missing statistics", addressString)
				m.Requester.QueryUnicast(addr, "GET statistics")
			}
		}
	}
}
