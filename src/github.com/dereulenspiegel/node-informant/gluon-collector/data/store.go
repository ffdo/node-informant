package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/gorilla/mux"
)

const TimeFormat string = "2006-01-02T15:04:05"

type NodeStatusInfo struct {
	Firstseen string
	Lastseen  string
	Online    bool
	Gateway   bool
}

type NodeinfoStore interface {
	GetNodeinfo(nodeId string) (NodeInfo, error)
	GetStatistics(nodeId string) (StatisticsStruct, error)
	GetNodeinfos() []NodeInfo
	GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error)
	LoadNodesFromFile(path string) error
	UpdateNodesJson()
}

type SimpleInMemoryStore struct {
	nodeinfos       map[string]NodeInfo
	statistics      map[string]StatisticsStruct
	statusInfo      map[string]NodeStatusInfo
	nodesJsonPath   string
	cachedNodesJson string
	gatewayList     map[string]bool
}

func NewSimpleInMemoryStore() *SimpleInMemoryStore {
	return &SimpleInMemoryStore{
		nodeinfos:  make(map[string]NodeInfo),
		statistics: make(map[string]StatisticsStruct),
		statusInfo: make(map[string]NodeStatusInfo),
	}
}

func (s *SimpleInMemoryStore) LoadNodesFromFile(path string) error {
	s.nodesJsonPath = path
	nodesFile, err := os.Open(path)
	defer nodesFile.Close()
	if err != nil {
		return err
	}
	jsonParser := json.NewDecoder(nodesFile)
	nodesJson := &NodesJson{}
	if err = jsonParser.Decode(nodesJson); err != nil {
		return err
	}
	for nodeId, nodeJsonInfo := range nodesJson.Nodes {
		nodeInfos := nodeJsonInfo.Nodeinfo
		nodeStats := nodeJsonInfo.Statistics
		nodeStatus := NodeStatusInfo{
			Firstseen: nodeJsonInfo.Firstseen,
			Lastseen:  nodeJsonInfo.Lastseen,
			Online:    nodeJsonInfo.Flags.Online,
			Gateway:   nodeJsonInfo.Flags.Gateway,
		}
		s.nodeinfos[nodeId] = nodeInfos
		s.statistics[nodeId] = nodeStats
		s.statusInfo[nodeId] = nodeStatus
	}
	return nil
}

func (s *SimpleInMemoryStore) updateNodeStatusInfo(response ParsedResponse) {
	info, exists := s.statusInfo[response.NodeId()]
	now := time.Now().Format(TimeFormat)
	_, isGw := s.gatewayList[response.NodeId()]
	if exists {
		info.Lastseen = now
		info.Online = true
	} else {
		info = NodeStatusInfo{
			Firstseen: now,
			Lastseen:  now,
			Online:    true,
			Gateway:   isGw,
		}
	}
	s.statusInfo[response.NodeId()] = info
}

func (s *SimpleInMemoryStore) GetNodeStatusInfo(nodeId string) (status NodeStatusInfo, err error) {
	stat, exists := s.statusInfo[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s has no status info", nodeId)
	}
	status = stat
	return
}

func (s *SimpleInMemoryStore) GetNodeinfo(nodeId string) (info NodeInfo, err error) {
	nodeinfo, exists := s.nodeinfos[nodeId]
	info = nodeinfo
	if !exists {
		err = fmt.Errorf("NodeId %s does not exist", nodeId)
		return
	}
	return
}

func (s *SimpleInMemoryStore) GetNodeinfos() []NodeInfo {
	nodeinfos := make([]NodeInfo, len(s.nodeinfos))
	counter := 0
	for _, nodeinfo := range s.nodeinfos {
		nodeinfos[counter] = nodeinfo
		counter++
	}
	return nodeinfos
}

func (s *SimpleInMemoryStore) GetStatistics(nodeId string) (statistics StatisticsStruct, err error) {
	stats, exists := s.statistics[nodeId]
	statistics = stats
	if !exists {
		err = fmt.Errorf("NodeId %s has no statistics", nodeId)
	}
	return
}

func (s *SimpleInMemoryStore) Routes() []httpserver.Route {
	var memoryStoreRoutes = []httpserver.Route{
		httpserver.Route{"NodeInfo", "GET", "/nodeinfos/{nodeid}", s.GetNodeInfoRest},
		httpserver.Route{"NodeInfos", "GET", "/nodeinfos", s.GetNodeInfosRest},
		httpserver.Route{"NodeStatistics", "GET", "/statistics/{nodeid}", s.GetNodeStatisticsRest},
		httpserver.Route{"NodesJson", "GET", "/nodes.json", s.GetNodesJsonRest},
	}
	return memoryStoreRoutes
}

func (s *SimpleInMemoryStore) GetNodeStatisticsRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	stats, err := s.GetStatistics(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}

func (s *SimpleInMemoryStore) GetNodeInfosRest(w http.ResponseWriter, r *http.Request) {
	nodeinfos := s.GetNodeinfos()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nodeinfos)
}

func (s *SimpleInMemoryStore) GetNodeInfoRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	nodeinfo, err := s.GetNodeinfo(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nodeinfo)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}

func (s *SimpleInMemoryStore) GetNodesJsonRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s.cachedNodesJson))
}

func (s *SimpleInMemoryStore) GetNodesJson() NodesJson {
	timestamp := time.Now().Format(TimeFormat)
	nodes := make(map[string]NodesJsonNode)
	for nodeId, nodeInfo := range s.nodeinfos {
		stats := s.statistics[nodeId]
		status := s.statusInfo[nodeId]
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

func (s *SimpleInMemoryStore) UpdateNodesJson() {
	nodesJson := s.GetNodesJson()

	data, err := json.Marshal(&nodesJson)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  s.nodesJsonPath,
		}).Error("Error encoding json")
		return
	}
	s.cachedNodesJson = string(data)
	if s.nodesJsonPath == "" {
		return
	}
	err = ioutil.WriteFile(s.nodesJsonPath, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  s.nodesJsonPath,
		}).Error("Error writing nodes.json")
	}
}

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
