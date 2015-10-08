package main

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

// MissingUpdater has the simple jop if iterating through all available NodeInfos
// and check whether we have statistics and neighbour infos for them. If they are
// missing these nodes are queried by unicast on the link local und global unicast
// IPv6 addresses for the missing information.
// This makes more sense if data in the data store can expire (this is a TODO for
// the BoltStore).
type MissingUpdater struct {
	Requester *announced.Requester
	Store     data.Nodeinfostore
}

// Query for the missing mesh neighbour information.
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

// Query for missing statistics.
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
