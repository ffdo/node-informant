package meshviewer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

const TimeFormat string = "2006-01-02T15:04:05"

type NodeFlags struct {
	Gateway bool `json:"gateway"`
	Online  bool `json:"online"`
}
type NodesJsonNode struct {
	Nodeinfo   data.NodeInfo     `json:"nodeinfo"`
	Statistics *StatisticsStruct `json:"statistics"`
	Flags      NodeFlags         `json:"flags"`
	Lastseen   string            `json:"lastseen"`
	Firstseen  string            `json:"firstseen"`
}

type NodesJson struct {
	Timestamp string                   `json:"timestamp"`
	Version   int                      `json:"version"`
	Nodes     map[string]NodesJsonNode `json:"nodes"`
}

type NodesJsonGenerator struct {
	Store           *data.SimpleInMemoryStore
	CachedNodesJson string
}

func (n *NodesJsonGenerator) Routes() []httpserver.Route {
	var nodesRoutes = []httpserver.Route{
		httpserver.Route{"NodesJson", "GET", "/nodes.json", n.GetNodesJsonRest},
	}
	return nodesRoutes
}

func (n *NodesJsonGenerator) GetNodesJsonRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(n.CachedNodesJson))
}

func convertToMeshviewerStatistics(in *data.StatisticsStruct) *StatisticsStruct {
	return &StatisticsStruct{
		Clients:     in.Clients.Total,
		Gateway:     in.Gateway,
		Loadavg:     in.LoadAverage,
		MemoryUsage: (float64(in.Memory.Free) / float64(in.Memory.Total)),
		RootfsUsage: in.RootFsUsage,
		Traffic:     &in.Traffic,
		Uptime:      in.Uptime,
	}
}

func (n *NodesJsonGenerator) GetNodesJson() NodesJson {
	timestamp := time.Now().Format(TimeFormat)
	nodes := make(map[string]NodesJsonNode)
	for nodeId, nodeInfo := range n.Store.Nodeinfos {
		var stats *StatisticsStruct = nil
		if storedStats, statsExist := n.Store.Statistics[nodeId]; statsExist {
			stats = convertToMeshviewerStatistics(storedStats)
		} else {
			stats = &StatisticsStruct{}
		}
		status := n.Store.StatusInfo[nodeId]
		flags := NodeFlags{
			Online:  status.Online,
			Gateway: status.Gateway,
		}
		node := NodesJsonNode{
			Nodeinfo:   nodeInfo,
			Statistics: stats,
			Lastseen:   status.Lastseen,
			Firstseen:  status.Firstseen,
			Flags:      flags,
		}
		nodes[nodeId] = node
	}
	nodesJson := NodesJson{
		Timestamp: timestamp,
		Version:   1,
		Nodes:     nodes,
	}
	return nodesJson
}

func (n *NodesJsonGenerator) UpdateNodesJson() {
	nodesJson := n.GetNodesJson()

	data, err := json.Marshal(&nodesJson)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"path":   n.Store.NodesJsonPath,
			"value":  err.(*json.UnsupportedValueError).Value,
			"string": err.(*json.UnsupportedValueError).Str,
		}).Errorf("Error encoding nodes.json")
		return
	}
	n.CachedNodesJson = string(data)
	if n.Store.NodesJsonPath == "" {
		return
	}
	err = ioutil.WriteFile(n.Store.NodesJsonPath, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  n.Store.NodesJsonPath,
		}).Error("Error writing nodes.json")
	}
}
