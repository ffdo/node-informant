package prometheus

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	stat "github.com/prometheus/client_golang/prometheus"
)

var (
	TotalClientCounter = stat.NewGauge(stat.GaugeOpts{
		Name: "total_clients",
		Help: "Total number of connected clients",
	})

	TotalNodes = stat.NewGauge(stat.GaugeOpts{
		Name: "total_nodes",
		Help: "Total number of Nodes",
	})

	TotalNodeTrafficRx = stat.NewCounter(stat.CounterOpts{
		Name: "total_traffic_rx",
		Help: "Total accumulated received traffic as reported by Nodes",
	})

	TotalNodeTrafficTx = stat.NewCounter(stat.CounterOpts{
		Name: "total_traffic_tx",
		Help: "Total accumulated transmitted traffic as reported by Nodes",
	})

	TotalNodeMgmtTrafficRx = stat.NewCounter(stat.CounterOpts{
		Name: "total_traffic_mgmt_rx",
		Help: "Total accumulated received management traffic as reported by Nodes",
	})

	TotalNodeMgmtTrafficTx = stat.NewCounter(stat.CounterOpts{
		Name: "total_traffic_mgmt_tx",
		Help: "Total accumulated transmitted management traffic as reported by Nodes",
	})

	OnlineNodes = stat.NewGauge(stat.GaugeOpts{
		Name: "node_online",
		Help: "All online nodes",
	})

	NodeMetricsMap = make(map[string]*NodeMetrics)
)

type NodeMetrics struct {
	Clients stat.Gauge
	Uptime  stat.Counter
	Traffic *stat.CounterVec
	NodeId  string
}

func GetNodeMetrics(nodeId string) *NodeMetrics {
	if m, exists := NodeMetricsMap[nodeId]; exists {
		return m
	}
	m := &NodeMetrics{
		NodeId: nodeId,
	}
	m.Uptime = stat.NewCounter(stat.CounterOpts{
		Name: fmt.Sprintf("node_%s_uptime", nodeId),
		Help: fmt.Sprintf("Uptime of node %s"),
	})
	m.Clients = stat.NewGauge(stat.GaugeOpts{
		Name: fmt.Sprintf("node_%s_clients", nodeId),
		Help: fmt.Sprintf("Clients connected to node %s", nodeId),
	})
	m.Traffic = stat.NewCounterVec(stat.CounterOpts{
		Name: fmt.Sprintf("node_%s_traffic", nodeId),
		Help: fmt.Sprintf("Measured traffic in bytes on %s", nodeId),
	}, []string{"type", "direction"})
	stat.Register(m.Uptime)
	stat.Register(m.Clients)
	stat.Register(m.Traffic)
	NodeMetricsMap[nodeId] = m
	return m
}

func init() {
	stat.MustRegister(TotalClientCounter)
	stat.MustRegister(TotalNodes)
	stat.MustRegister(TotalNodeTrafficRx)
	stat.MustRegister(TotalNodeTrafficTx)
	stat.MustRegister(TotalNodeMgmtTrafficRx)
	stat.MustRegister(TotalNodeMgmtTrafficTx)
	stat.MustRegister(OnlineNodes)
}

func initTotalClientsGauge(store data.Nodeinfostore) {
	TotalClientCounter.Set(0.0)
	var totalClients int = 0
	for _, stats := range store.GetAllStatistics() {
		status, err := store.GetNodeStatusInfo(stats.NodeId)
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err,
				"nodeId": stats.NodeId,
			}).Warn("Didn't found node status in store")
		}
		if status.Online {
			totalClients = totalClients + stats.Clients.Total
			log.Debugf("Adding %d clients", stats.Clients.Total)
			TotalClientCounter.Add(float64(stats.Clients.Total))
		} else {
			log.Debugf("Node %s was offline", status.NodeId)
		}
	}
	log.Debugf("Initialised prometheus with %d clients", totalClients)
}

func initTrafficCounter(store data.Nodeinfostore) {
	TotalNodeTrafficRx.Set(0.0)
	TotalNodeTrafficTx.Set(0.0)
	TotalNodeMgmtTrafficRx.Set(0.0)
	TotalNodeMgmtTrafficTx.Set(0.0)

	for _, stats := range store.GetAllStatistics() {
		TotalNodeTrafficRx.Add(float64(stats.Traffic.Rx.Bytes))
		TotalNodeTrafficTx.Add(float64(stats.Traffic.Tx.Bytes))
		TotalNodeMgmtTrafficRx.Add(float64(stats.Traffic.MgmtRx.Bytes))
		TotalNodeMgmtTrafficTx.Add(float64(stats.Traffic.MgmtTx.Bytes))
	}
}

func decrementOnlineNodes(nodeId string) {
	OnlineNodes.Dec()
}

func initOnlineNodesGauge(store data.Nodeinfostore) {
	OnlineNodes.Set(0.0)
	for _, status := range store.GetNodeStatusInfos() {
		if status.Online {
			OnlineNodes.Inc()
		}
	}
	store.NotifyNodeOffline(decrementOnlineNodes)
}

func ProcessStoredValues(store data.Nodeinfostore) {
	TotalNodes.Set(float64(len(store.GetNodeStatusInfos())))
	initTotalClientsGauge(store)
	initTrafficCounter(store)
	initOnlineNodesGauge(store)
}
