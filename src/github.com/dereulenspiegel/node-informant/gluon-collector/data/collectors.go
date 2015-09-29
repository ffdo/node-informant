package data

type GatewayCollector struct {
	Store *SimpleInMemoryStore
}

func (g *GatewayCollector) Process(in chan ParsedResponse) chan ParsedResponse {
	out := make(chan ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(StatisticsStruct)
				gateway := statistics.Gateway
				if gateway != "" {
					_, exists := g.Store.gatewayList[response.NodeId()]
					if !exists {
						g.Store.gatewayList[response.NodeId()] = true
					}
				}
			}
			out <- response
		}
	}()
	return out
}

type NodeinfoCollector struct {
	Store *SimpleInMemoryStore
}

func (n *NodeinfoCollector) Process(in chan ParsedResponse) chan ParsedResponse {
	out := make(chan ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "nodeinfo" {
				nodeinfo := response.ParsedData().(NodeInfo)
				n.Store.nodeinfos[nodeinfo.NodeId] = nodeinfo
			}
			out <- response
		}
	}()
	return out
}

type StatisticsCollector struct {
	Store *SimpleInMemoryStore
}

func (s *StatisticsCollector) Process(in chan ParsedResponse) chan ParsedResponse {
	out := make(chan ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				statistics := response.ParsedData().(StatisticsStruct)
				s.Store.statistics[statistics.NodeId] = statistics
			}
			out <- response
		}
	}()
	return out
}

type NeighbourInfoCollector struct {
	Store *SimpleInMemoryStore
}

func (n *NeighbourInfoCollector) Process(in chan ParsedResponse) chan ParsedResponse {
	out := make(chan ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "neighbours" {
				neighbours := response.ParsedData().(NeighbourStruct)
				n.Store.neighbourInfos[neighbours.NodeId] = neighbours
			}
			out <- response
		}
	}()
	return out
}
