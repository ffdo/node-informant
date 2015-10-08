package pipeline

import (
	"time"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

// GatewayCollector inspects all received statistics and stores the mac addresses
// of gateways to the data store.
type GatewayCollector struct {
	Store data.Nodeinfostore
}

func (g *GatewayCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(*data.StatisticsStruct)
				gateway := statistics.Gateway
				if gateway != "" {
					g.Store.PutGateway(gateway)
				}
			}
			out <- response
		}

	}()
	return out
}

// NodeinfoCollector inspects all ParsedResponses containing general information
// about a node and stores this to the data store.
type NodeinfoCollector struct {
	Store data.Nodeinfostore
}

func (n *NodeinfoCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "nodeinfo" {
				nodeinfo := response.ParsedData().(data.NodeInfo)
				n.Store.PutNodeInfo(nodeinfo)
			}
			out <- response
		}

	}()
	return out
}

// StatisticsCollector collects all ParsedResponses containing statistics information
// and stores them in the data store.
type StatisticsCollector struct {
	Store data.Nodeinfostore
}

func (s *StatisticsCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(*data.StatisticsStruct)
				s.Store.PutStatistics(*statistics)
			}
			out <- response
		}

	}()
	return out
}

// NeighbourInfoCollector inspects all ParsedResponses containing information about
// mesh neighbours and stores them to the data store.
type NeighbourInfoCollector struct {
	Store data.Nodeinfostore
}

func (n *NeighbourInfoCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "neighbours" {
				neighbours := response.ParsedData().(*data.NeighbourStruct)
				n.Store.PutNodeNeighbours(*neighbours)
			}
			out <- response
		}

	}()
	return out
}

const TimeFormat string = time.RFC3339

// StatusInfoCollector creates some meta data like Firstseen and Lastseen for every
// node. Everytime we receive a packet from a node, we assume that is online and also
// update the Lastseen value. If we have never seen a packet from this node before we
// also set the Firstseen value.
// TODO determine Gateway status.
type StatusInfoCollector struct {
	Store data.Nodeinfostore
}

func (s *StatusInfoCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			nodeId := response.NodeId()
			statusInfo, err := s.Store.GetNodeStatusInfo(nodeId)
			if err == nil {
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
			s.Store.PutNodeStatusInfo(nodeId, statusInfo)
			out <- response
		}

	}()
	return out
}
