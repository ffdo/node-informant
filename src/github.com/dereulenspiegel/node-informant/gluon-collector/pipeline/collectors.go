package pipeline

import "github.com/dereulenspiegel/node-informant/gluon-collector/data"

type GatewayCollector struct {
	Store *data.SimpleInMemoryStore
}

func (g *GatewayCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(data.StatisticsStruct)
				gateway := statistics.Gateway
				if gateway != "" {
					_, exists := g.Store.GatewayList[response.NodeId()]
					if !exists {
						g.Store.GatewayList[response.NodeId()] = true
					}
				}
			}
			out <- response
		}
	}()
	return out
}

type NodeinfoCollector struct {
	Store *data.SimpleInMemoryStore
}

func (n *NodeinfoCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "nodeinfo" {
				nodeinfo := response.ParsedData().(data.NodeInfo)
				n.Store.Nodeinfos[nodeinfo.NodeId] = nodeinfo
			}
			out <- response
		}
	}()
	return out
}

type StatisticsCollector struct {
	Store *data.SimpleInMemoryStore
}

func (s *StatisticsCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(data.StatisticsStruct)
				s.Store.Statistics[statistics.NodeId] = statistics
			}
			out <- response
		}
	}()
	return out
}

type NeighbourInfoCollector struct {
	Store *data.SimpleInMemoryStore
}

func (n *NeighbourInfoCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "neighbours" {
				neighbours := response.ParsedData().(data.NeighbourStruct)
				n.Store.NeighbourInfos[neighbours.NodeId] = neighbours
			}
			out <- response
		}
	}()
	return out
}
