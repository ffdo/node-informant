package collectors

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/prometheus"
)

// GatewayCollector inspects all received statistics and stores the mac addresses
// of gateways to the data store.
func GatewayCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	if statistics, ok := response.ParsedData().(*data.StatisticsStruct); ok {
		gateway := statistics.Gateway
		if gateway != "" {
			store.PutGateway(gateway)
		}
	}
}

// NodeinfoCollector inspects all ParsedResponses containing general information
// about a node and stores this to the data store.
func NodeinfoCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	if nodeinfo, ok := response.ParsedData().(data.NodeInfo); ok {
		store.PutNodeInfo(nodeinfo)
	}
}

// StatisticsCollector collects all ParsedResponses containing statistics information
// and stores them in the data store.
func StatisticsCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	if statistics, ok := response.ParsedData().(*data.StatisticsStruct); ok {
		store.PutStatistics(*statistics)
	}
}

// NeighbourInfoCollector inspects all ParsedResponses containing information about
// mesh neighbours and stores them to the data store.
func NeighbourInfoCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	if neighbours, ok := response.ParsedData().(*data.NeighbourStruct); ok {
		store.PutNodeNeighbours(*neighbours)
	}
}

const TimeFormat string = time.RFC3339

// StatusInfoCollector creates some meta data like Firstseen and Lastseen for every
// node. Everytime we receive a packet from a node, we assume that is online and also
// update the Lastseen value. If we have never seen a packet from this node before we
// also set the Firstseen value.
// TODO determine Gateway status.
func StatusInfoCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	nodeId := response.NodeId()
	statusInfo, err := store.GetNodeStatusInfo(nodeId)
	if err == nil {
		if !statusInfo.Online {
			prometheus.OnlineNodes.Inc()
			log.WithFields(log.Fields{
				"nodeid": nodeId,
			}).Info("Node is considered online again, after receiving any packet at all")
		}
		statusInfo.Online = true
		statusInfo.Lastseen = time.Now().Format(TimeFormat)
	} else {
		statusInfo = data.NodeStatusInfo{
			Online:    true,
			Firstseen: time.Now().Format(TimeFormat),
			Lastseen:  time.Now().Format(TimeFormat),
			Gateway:   false,
			NodeId:    nodeId,
		}
	}
	store.PutNodeStatusInfo(nodeId, statusInfo)

}
