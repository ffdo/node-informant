package prometheus

import (
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	stat "github.com/prometheus/client_golang/prometheus"
)

// NodeCountPipe simply increments the TotalNodes Gauge if we receive a response from
// a node we don't habe a NodeStatusInfo object for. This indicates that we didn't had
// contact with this node before and therefore it is new.
// It also increments the OnlineNodes Gauge by one in case the stored NodeStatusInfo
// indicates that this node wasn't online before.
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
				// New node. Also increment online count
				OnlineNodes.Inc()
			}
			out <- response
		}
	}()
	return out
}

type ReturnedNodeDetector struct {
	Store data.Nodeinfostore
}

func (r *ReturnedNodeDetector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			status, err := r.Store.GetNodeStatusInfo(response.NodeId())
			if err == nil && !status.Online {
				// Existing offline node came back online
				OnlineNodes.Inc()
			}
			out <- response
		}
	}()
	return out
}

// ClientCountPipe tries to determine the difference in clients per node
// between the currently received statistics and the received statistics and adds
// the difference to the TotalClientsCounter Gauge.
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

// TrafficCountPipe determines the difference between previous node traffic and current
// node traffic and increments the Total traffic counters by the difference.
// TODO: Have look whether CounterVec is a better choice than 4 different counters.
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

// NodeMetricCollector updates per node metrics based on received statistics responses.
type NodeMetricCollector struct{}

func (n *NodeMetricCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				stats := response.ParsedData().(*data.StatisticsStruct)
				NodesClients.WithLabelValues(response.NodeId()).Set(float64(stats.Clients.Total))
				NodesUptime.WithLabelValues(response.NodeId()).Set(stats.Uptime)
				if stats.Traffic != nil {
					NodesTrafficRx.WithLabelValues(response.NodeId(), "traffic").Set(float64(stats.Traffic.Rx.Bytes))
					NodesTrafficTx.WithLabelValues(response.NodeId(), "traffic").Set(float64(stats.Traffic.Tx.Bytes))
					NodesTrafficRx.WithLabelValues(response.NodeId(), "mgmt_traffic").Set(float64(stats.Traffic.MgmtRx.Bytes))
					NodesTrafficTx.WithLabelValues(response.NodeId(), "mgmt_traffic").Set(float64(stats.Traffic.MgmtTx.Bytes))
				}
			}
			out <- response
		}
	}()
	return out
}

// GetPrometheusProcessPipes returns all ProcessPipes necessary to keep Prometheus
// metrics up to date. In most cases the Prometheus pipes need to be added before
// all other pipes to the ProcessPipeline.
func GetPrometheusProcessPipes(store data.Nodeinfostore) []pipeline.ProcessPipe {
	out := make([]pipeline.ProcessPipe, 0, 10)
	out = append(out, &NodeCountPipe{Store: store})
	out = append(out, &ClientCountPipe{Store: store})
	out = append(out, &TrafficCountPipe{Store: store})
	out = append(out, &NodeMetricCollector{})
	out = append(out, &ReturnedNodeDetector{Store: store})
	return out
}
