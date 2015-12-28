package meshviewer

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

const TimeFormat string = time.RFC3339

type NodeFlags struct {
	Gateway bool `json:"gateway"`
	Online  bool `json:"online"`
	Uplink  bool `json:"uplink"`
}

// NodesJsonNode is the structure expected by mehsviewer to contain all information
// about a node.
type NodesJsonNode struct {
	Nodeinfo   data.NodeInfo     `json:"nodeinfo"`
	Statistics *StatisticsStruct `json:"statistics"`
	Flags      NodeFlags         `json:"flags"`
	Lastseen   string            `json:"lastseen"`
	Firstseen  string            `json:"firstseen"`
}

// NodesJson is the top level structure for nodes.json
type NodesJson struct {
	Timestamp string                   `json:"timestamp"`
	Version   int                      `json:"version"`
	Nodes     map[string]NodesJsonNode `json:"nodes"`
}

type NodesJsonV2 struct {
	Timestamp string          `json:"timestamp"`
	Version   int             `json:"version"`
	Nodes     []NodesJsonNode `json:"nodes"`
}

type NodesJsonGenerator struct {
	Store             data.Nodeinfostore
	CachedNodesJson   string
	meshviewerVersion int
}

func NewNodesJsonGenerator(store data.Nodeinfostore) *NodesJsonGenerator {
	meshviewerVersion := config.Global.UInt("meshviewer_version", 1)

	if meshviewerVersion < 1 && meshviewerVersion > 2 {
		log.Fatalf("Invalid meshviewer version %d", meshviewerVersion)
	}
	return &NodesJsonGenerator{meshviewerVersion: meshviewerVersion, Store: store}
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

// convertToMeshviewerStatistics takes care of converting statistics received by
// announced to something meshviewer can digest.
func convertToMeshviewerStatistics(in *data.StatisticsStruct) StatisticsStruct {
	return StatisticsStruct{
		Clients:     in.Clients.Total,
		Gateway:     in.Gateway,
		Loadavg:     in.LoadAverage,
		MemoryUsage: ((float64(in.Memory.Total) - float64(in.Memory.Free) - float64(in.Memory.Buffers) - float64(in.Memory.Cached)) / float64(in.Memory.Total)),
		RootfsUsage: in.RootFsUsage,
		Traffic:     in.Traffic,
		Uptime:      in.Uptime,
	}
}

func determineUplink(stats data.StatisticsStruct) bool {
	if stats.MeshVpn != nil {
		for _, group := range stats.MeshVpn.Groups {
			if group != nil {
				for _, peer := range group.Peers {
					if peer != nil && peer.Established > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}

// GetNodesJson fills a NodesJson struct with all information stored in the
// Nodeinfostore
func (n *NodesJsonGenerator) GetNodesJson() NodesJson {
	timestamp := time.Now().Format(TimeFormat)
	nodes := make(map[string]NodesJsonNode)
	for _, nodeInfo := range n.Store.GetNodeInfos() {
		nodeId := nodeInfo.NodeId
		status, _ := n.Store.GetNodeStatusInfo(nodeId)
		var stats StatisticsStruct
		flags := NodeFlags{
			Online:  status.Online,
			Gateway: status.Gateway,
		}
		if storedStats, err := n.Store.GetStatistics(nodeId); err == nil {
			if !status.Online {
				storedStats.Clients.Wifi = 0
				storedStats.Clients.Total = 0
			}
			flags.Uplink = determineUplink(storedStats)
			stats = convertToMeshviewerStatistics(&storedStats)
		} else {
			stats = StatisticsStruct{}
		}

		node := NodesJsonNode{
			Nodeinfo:   nodeInfo,
			Statistics: &stats,
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

// GetNodesJson fills a NodesJsonV2 struct with all information stored in the
// Nodeinfostore
func (n *NodesJsonGenerator) GetNodesJsonV2() NodesJsonV2 {
	timestamp := time.Now().Format(TimeFormat)
	nodeInfos := n.Store.GetNodeInfos()
	nodes := make([]NodesJsonNode, 0, len(nodeInfos))
	for _, nodeInfo := range nodeInfos {
		nodeId := nodeInfo.NodeId
		status, _ := n.Store.GetNodeStatusInfo(nodeId)
		var stats StatisticsStruct
		flags := NodeFlags{
			Online:  status.Online,
			Gateway: status.Gateway,
		}
		if storedStats, err := n.Store.GetStatistics(nodeId); err == nil {
			if !status.Online {
				storedStats.Clients.Wifi = 0
				storedStats.Clients.Total = 0
			}
			flags.Uplink = determineUplink(storedStats)
			stats = convertToMeshviewerStatistics(&storedStats)
		} else {
			stats = StatisticsStruct{}
		}
		node := NodesJsonNode{
			Nodeinfo:   nodeInfo,
			Statistics: &stats,
			Lastseen:   status.Lastseen,
			Firstseen:  status.Firstseen,
			Flags:      flags,
		}
		nodes = append(nodes, node)
	}
	nodesJson := NodesJsonV2{
		Timestamp: timestamp,
		Version:   2,
		Nodes:     nodes,
	}
	return nodesJson
}

// UpdateNodesJson creates a new json string from a freshly generated NodesJson
// and caches it so, that the REST handlers can simply send the cached string.
func (n *NodesJsonGenerator) UpdateNodesJson() {

	var nodeData interface{}
	switch n.meshviewerVersion {
	case 1:
		nodeData = n.GetNodesJson()
	case 2:
		nodeData = n.GetNodesJsonV2()
	}

	data, err := json.Marshal(&nodeData)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"value":  err.(*json.UnsupportedValueError).Value,
			"string": err.(*json.UnsupportedValueError).Str,
		}).Errorf("Error marshalling nodes.json")
	} else {
		n.CachedNodesJson = string(data)
	}
}
