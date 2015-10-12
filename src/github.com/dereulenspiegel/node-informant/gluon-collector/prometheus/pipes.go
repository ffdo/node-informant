package prometheus

import (
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	stat "github.com/prometheus/client_golang/prometheus"
)

type NodeCountPipe struct {
	Store data.Nodeinfostore
}

func (n *NodeCountPipe) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			status, err := n.Store.GetNodeStatusInfo(response.NodeId())
			if err != nil {
				TotalNodes.Inc()
				// New node. Also increment online count
				OnlineNodes.Inc()
			} else if status.Online == false {
				// Existing offline node came back online
				OnlineNodes.Inc()
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
				var addValue float64
				if err == nil {
					addValue = float64(newStats.Clients.Total - oldStats.Clients.Total)
				} else {
					addValue = float64(newStats.Clients.Total)
				}
				TotalClientCounter.Add(addValue)
			}
			out <- response
		}
	}()
	return out
}

type TrafficCountPipe struct {
	Store data.Nodeinfostore
}

func collectTrafficBytes(counter stat.Counter, oldTraffic, newTraffic *data.TrafficObject) {
	var value float64
	if oldTraffic != nil {
		value = float64(newTraffic.Bytes - oldTraffic.Bytes)
	} else {
		value = float64(newTraffic.Bytes)
	}
	counter.Add(value)
}

func (t *TrafficCountPipe) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				newStats, _ := response.ParsedData().(*data.StatisticsStruct)
				oldStats, _ := t.Store.GetStatistics(response.NodeId())

				if oldStats.Traffic == nil {
					oldStats.Traffic = &data.TrafficStruct{}
				}
				collectTrafficBytes(TotalNodeTrafficTx, oldStats.Traffic.Tx, newStats.Traffic.Tx)
				collectTrafficBytes(TotalNodeTrafficRx, oldStats.Traffic.Rx, newStats.Traffic.Rx)
				collectTrafficBytes(TotalNodeMgmtTrafficRx, oldStats.Traffic.MgmtRx, newStats.Traffic.MgmtRx)
				collectTrafficBytes(TotalNodeMgmtTrafficTx, oldStats.Traffic.MgmtTx, newStats.Traffic.MgmtTx)
			}
			out <- response
		}
	}()
	return out
}

type NodeMetricCollector struct{}

func (n *NodeMetricCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				stats := response.ParsedData().(*data.StatisticsStruct)
				metrics := GetNodeMetrics(response.NodeId())
				metrics.Clients.Set(float64(stats.Clients.Total))
				metrics.Uptime.Set(stats.Uptime)
				if stats.Traffic != nil {
					metrics.Traffic.WithLabelValues("traffic", "rx").Set(float64(stats.Traffic.Rx.Bytes))
					metrics.Traffic.WithLabelValues("traffic", "tx").Set(float64(stats.Traffic.Tx.Bytes))
					metrics.Traffic.WithLabelValues("mgmt_traffic", "rx").Set(float64(stats.Traffic.MgmtRx.Bytes))
					metrics.Traffic.WithLabelValues("mgmt_traffic", "tx").Set(float64(stats.Traffic.MgmtTx.Bytes))
				}
			}
			out <- response
		}
	}()
	return out
}

func GetPrometheusProcessPipes(store data.Nodeinfostore) []pipeline.ProcessPipe {
	out := make([]pipeline.ProcessPipe, 0, 10)
	out = append(out, &NodeCountPipe{Store: store})
	out = append(out, &ClientCountPipe{Store: store})
	out = append(out, &TrafficCountPipe{Store: store})
	out = append(out, &NodeMetricCollector{})
	return out
}
