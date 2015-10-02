package pipeline

import (
	"time"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

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
					/*_, exists := g.Store.GatewayList[response.NodeId()]
					if !exists {
						g.Store.GatewayList[response.NodeId()] = true
					}*/
				}
			}
			out <- response
		}
	}()
	return out
}

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
				//n.Store.Nodeinfos[nodeinfo.NodeId] = nodeinfo
			}
			out <- response
		}
	}()
	return out
}

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
				//s.Store.Statistics[statistics.NodeId] = statistics
			}
			out <- response
		}
	}()
	return out
}

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
				//n.Store.NeighbourInfos[neighbours.NodeId] = neighbours
			}
			out <- response
		}
	}()
	return out
}

const TimeFormat string = "2006-01-02T15:04:05"

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
				}
			}
			s.Store.PutNodeStatusInfo(nodeId, statusInfo)
			out <- response
		}
	}()
	return out
}
