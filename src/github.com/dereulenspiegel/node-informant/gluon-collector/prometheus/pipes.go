package prometheus

import (
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type NodeCountPipe struct {
	Store data.Nodeinfostore
}

func (n *NodeCountPipe) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			_, err := n.Store.GetNodeStatusInfo(response.NodeId())
			if err != nil {
				TotalNodes.Inc()
			}
			out <- response
		}
	}()
	return out
}

type ClientCountPipe struct {
	Store data.Nodeinfostore
}

func (c *ClientCountPipe) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				newStats, _ := response.ParsedData().(*data.StatisticsStruct)
				oldStats, err := c.Store.GetStatistics(response.NodeId())
				if err != nil {
					clientDiff := oldStats.Clients.Total - newStats.Clients.Total
					if clientDiff > 0 {
						TotalClientCounter.Add(float64(clientDiff))
					} else {
						TotalClientCounter.Sub(float64(clientDiff))
					}
				} else {
					TotalClientCounter.Add(float64(newStats.Clients.Total))
				}
			}
			out <- response
		}
	}()
	return out
}
