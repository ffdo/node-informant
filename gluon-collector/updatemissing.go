package main

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/ffdo/node-informant/announced"
	"github.com/ffdo/node-informant/gluon-collector/data"
)

// MissingUpdater has the simple jop if iterating through all available NodeInfos
// and check whether we have statistics and neighbour infos for them. If they are
// missing these nodes are queried by unicast on the link local und global unicast
// IPv6 addresses for the missing information.
// This makes more sense if data in the data store can expire (this is a TODO for
// the BoltStore).
type MissingUpdater struct {
	Requester announced.AnnouncedPacketReceiver
	Store     data.Nodeinfostore
}

func (m *MissingUpdater) CheckNodeUnicast(nodeId string) {
	nodeinfo, err := m.Store.GetNodeInfo(nodeId)
	if err != nil {
		// TODO log this, this shouldn't happen
		return
	}
	m.UpdateMissingNeighbours(nodeinfo)
	m.UpdateMissingStatistics(nodeinfo)
}

// Query for the missing mesh neighbour information.
func (m *MissingUpdater) UpdateMissingNeighbours(nodeinfo data.NodeInfo) {
	log.WithFields(log.Fields{
		"nodeid": nodeinfo.NodeId,
	}).Info("Updating missing neighbour infos")

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

// Query for missing statistics.
func (m *MissingUpdater) UpdateMissingStatistics(nodeinfo data.NodeInfo) {
	log.WithFields(log.Fields{
		"nodeid": nodeinfo.NodeId,
	}).Info("Updating missing statistics")

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
