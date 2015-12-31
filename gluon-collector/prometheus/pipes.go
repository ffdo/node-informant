package prometheus

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/gluon-collector/config"
	"github.com/ffdo/node-informant/gluon-collector/data"
	stat "github.com/prometheus/client_golang/prometheus"
)

// GetPrometheusProcessPipes returns all ProcessPipes necessary to keep Prometheus
// metrics up to date. In most cases the Prometheus pipes need to be added before
// all other pipes to the ProcessPipeline.
func GetPrometheusProcessPipes() []data.ParsedResponseReader {
	return []data.ParsedResponseReader{
		NodeCounter,
		ReturnedNodeDetector,
		ClientCounter,
		TrafficCounter,
		NodeMetricCollector,
	}
}

// NodeCounter simply increments the TotalNodes Gauge if we receive a response from
// a node we don't habe a NodeStatusInfo object for. This indicates that we didn't had
// contact with this node before and therefore it is new.
// It also increments the OnlineNodes Gauge by one in case the stored NodeStatusInfo
// indicates that this node wasn't online before.
func NodeCounter(store data.Nodeinfostore, response data.ParsedResponse) {
	if _, ok := response.ParsedData().(data.NodeInfo); !ok {
		return
	}
	_, err := store.GetNodeStatusInfo(response.NodeId())
	if err != nil {
		log.WithFields(log.Fields{
			"nodeid": response.NodeId(),
		}).Info("Discovered new node")
		OnlineNodes.Inc()
		TotalNodes.Inc()
	}
}

// ReturnedNodeDetector handles incrementing the online node metric for returning
// nodes.
func ReturnedNodeDetector(store data.Nodeinfostore, response data.ParsedResponse) {
	if _, ok := response.ParsedData().(data.NodeInfo); !ok {
		return
	}
	status, err := store.GetNodeStatusInfo(response.NodeId())
	if err == nil && !status.Online {
		// Existing offline node came back online
		log.Debugf("Node %s came back online", response.NodeId())
		OnlineNodes.Inc()
	}
}

// ClientCounter tries to determine the difference in clients per node
// between the currently received statistics and the received statistics and adds
// the difference to the TotalClientsCounter Gauge.
func ClientCounter(store data.Nodeinfostore, response data.ParsedResponse) {
	newStats, ok := response.ParsedData().(*data.StatisticsStruct)
	if !ok {
		return
	}

	oldStats, err := store.GetStatistics(response.NodeId())
	var addValue float64
	if err == nil {
		addValue = float64(newStats.Clients.Total - oldStats.Clients.Total)
	} else {
		addValue = float64(newStats.Clients.Total)
	}
	TotalClientCounter.Add(addValue)
}

// TrafficCounter determines the difference between previous node traffic and current
// node traffic and increments the Total traffic counters by the difference.
// TODO: Have look whether CounterVec is a better choice than 4 different counters.

func TrafficCounter(store data.Nodeinfostore, response data.ParsedResponse) {
	newStats, ok := response.ParsedData().(*data.StatisticsStruct)
	if !ok {
		return
	}

	oldStats, _ := store.GetStatistics(response.NodeId())

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

// NodeMetricCollector updates per node metrics based on received statistics responses.
func NodeMetricCollector(store data.Nodeinfostore, response data.ParsedResponse) {
	stats, ok := response.ParsedData().(*data.StatisticsStruct)
	if !ok {
		return
	}
	var labels []string
	nodeinfo, err := store.GetNodeInfo(response.NodeId())
	if err != nil {
		// Extended labels are not configured
		if _, err := config.Global.Get("prometheus"); err != nil {
			labels = []string{response.NodeId()}
		} else {
			log.WithFields(log.Fields{
				"nodeid": response.NodeId(),
			}).Errorf("Can't retrieve node infos to get the hostname")
			return
		}
	} else {
		labels = getLabels(nodeinfo)
	}

	NodesClients.WithLabelValues(labels...).Set(float64(stats.Clients.Total))
	NodesUptime.WithLabelValues(labels...).Set(stats.Uptime)
	if stats.Traffic != nil {
		if stats.Traffic.Rx != nil {
			NodesTrafficRx.WithLabelValues(append(labels, "traffic")...).Set(float64(stats.Traffic.Rx.Bytes))
		}
		if stats.Traffic.Tx != nil {
			NodesTrafficTx.WithLabelValues(append(labels, "traffic")...).Set(float64(stats.Traffic.Tx.Bytes))
		}
		if stats.Traffic.MgmtRx != nil {
			NodesTrafficRx.WithLabelValues(append(labels, "mgmt_traffic")...).Set(float64(stats.Traffic.MgmtRx.Bytes))
		}
		if stats.Traffic.MgmtTx != nil {
			NodesTrafficTx.WithLabelValues(append(labels, "mgmt_traffic")...).Set(float64(stats.Traffic.MgmtTx.Bytes))
		}
	}
}

func getLabels(nodeinfo data.NodeInfo) (labels []string) {
	labels = make([]string, 0, 5)
	labels = append(labels, nodeinfo.NodeId)

	prometheusCfg, err := config.Global.Get("prometheus")
	if err == nil {
		if prometheusCfg.UBool("namelabel", false) {
			labels = append(labels, nodeinfo.Hostname)
		}
		if prometheusCfg.UBool("sitecodelabel", false) {
			labels = append(labels, nodeinfo.System.SiteCode)
		}
	}
	return labels
}
