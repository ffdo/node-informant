package prometheus

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/gluon-collector/config"
	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/pipeline"
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
			if response.Type() == "nodeinfo" {
				_, err := n.Store.GetNodeStatusInfo(response.NodeId())
				if err != nil {
					log.WithFields(log.Fields{
						"nodeid": response.NodeId(),
					}).Info("Discovered new node")
					OnlineNodes.Inc()
					TotalNodes.Inc()
				}
			}
			out <- response
		}
	}()
	return out
}

// ReturnedNodeDetector handles incrementing the online node metric for returning
// nodes.
type ReturnedNodeDetector struct {
	Store data.Nodeinfostore
}

func (r *ReturnedNodeDetector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "nodeinfo" {
				status, err := r.Store.GetNodeStatusInfo(response.NodeId())
				if err == nil && !status.Online {
					// Existing offline node came back online
					log.Debugf("Node %s came back online", response.NodeId())
					OnlineNodes.Inc()
				}
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
	if newTraffic == nil {
		// If a client statistics has no traffic information, don't collect anything
		return
	}
	var value float64
	if oldTraffic != nil {
		value = float64(newTraffic.Bytes - oldTraffic.Bytes)
	} else {
		value = float64(newTraffic.Bytes)
	}
	if value > 0 {
		counter.Add(value)
	} else if newTraffic.Bytes > 0 {
		counter.Add(newTraffic.Bytes)
	} else {
		log.WithFields(log.Fields{
			"newTraffic": *newTraffic,
			"oldTraffic": oldTraffic,
			"value":      value,
		}).Errorf("New traffic value was smaller than the old value and the new traffic value even seemed to be negative")
	}
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
				if newStats.Traffic == nil {
					newStats.Traffic = &data.TrafficStruct{}
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
type NodeMetricCollector struct {
	Store data.Nodeinfostore
}

func getLabels(nodeinfo data.NodeInfo, defaultLabels ...string) []string {
	labels := make([]string, 0, 5)
	labels = append(labels, nodeinfo.NodeId)
	prometheusCfg, err := config.Global.Get("prometheus")
	if err != nil {
		return append(labels, defaultLabels...)
	}
	if prometheusCfg.UBool("namelabel", false) {
		labels = append(labels, nodeinfo.Hostname)
	}
	if prometheusCfg.UBool("sitecodelabel", false) {
		labels = append(labels, nodeinfo.System.SiteCode)
	}
	return append(labels, defaultLabels...)
}

func (n *NodeMetricCollector) Process(in chan data.ParsedResponse) chan data.ParsedResponse {
	out := make(chan data.ParsedResponse)
	go func() {
		for response := range in {
			if response.Type() == "statistics" {
				stats := response.ParsedData().(*data.StatisticsStruct)
				nodeinfo, err := n.Store.GetNodeInfo(response.NodeId())
				if err != nil {
					// Extended labels are not configured
					if _, err := config.Global.Get("prometheus"); err != nil {
						// FIXME This is bad code duplication, we should find a more elegant way
						NodesClients.WithLabelValues(response.NodeId()).Set(float64(stats.Clients.Total))
						NodesUptime.WithLabelValues(response.NodeId()).Set(stats.Uptime)
						if stats.Traffic != nil {
							if stats.Traffic.Rx != nil {
								NodesTrafficRx.WithLabelValues(response.NodeId(), "traffic").Set(float64(stats.Traffic.Rx.Bytes))
							}
							if stats.Traffic.Tx != nil {
								NodesTrafficTx.WithLabelValues(response.NodeId(), "traffic").Set(float64(stats.Traffic.Tx.Bytes))
							}
							if stats.Traffic.MgmtRx != nil {
								NodesTrafficRx.WithLabelValues(response.NodeId(), "mgmt_traffic").Set(float64(stats.Traffic.MgmtRx.Bytes))
							}
							if stats.Traffic.MgmtTx != nil {
								NodesTrafficTx.WithLabelValues(response.NodeId(), "mgmt_traffic").Set(float64(stats.Traffic.MgmtTx.Bytes))
							}
						}
					} else {
						log.WithFields(log.Fields{
							"nodeid": response.NodeId(),
						}).Errorf("Can't retrieve node infos to get the hostname")
					}
				} else {
					NodesClients.WithLabelValues(getLabels(nodeinfo)...).Set(float64(stats.Clients.Total))
					NodesUptime.WithLabelValues(getLabels(nodeinfo)...).Set(stats.Uptime)
					if stats.Traffic != nil {
						if stats.Traffic.Rx != nil {
							NodesTrafficRx.WithLabelValues(getLabels(nodeinfo, "traffic")...).Set(float64(stats.Traffic.Rx.Bytes))
						}
						if stats.Traffic.Tx != nil {
							NodesTrafficTx.WithLabelValues(getLabels(nodeinfo, "traffic")...).Set(float64(stats.Traffic.Tx.Bytes))
						}
						if stats.Traffic.MgmtRx != nil {
							NodesTrafficRx.WithLabelValues(getLabels(nodeinfo, "mgmt_traffic")...).Set(float64(stats.Traffic.MgmtRx.Bytes))
						}
						if stats.Traffic.MgmtTx != nil {
							NodesTrafficTx.WithLabelValues(getLabels(nodeinfo, "mgmt_traffic")...).Set(float64(stats.Traffic.MgmtTx.Bytes))
						}
					}
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
	return []pipeline.ProcessPipe{
		&NodeCountPipe{Store: store},
		&ReturnedNodeDetector{Store: store},
		&ClientCountPipe{Store: store},
		&TrafficCountPipe{Store: store},
		&NodeMetricCollector{Store: store},
	}
}
