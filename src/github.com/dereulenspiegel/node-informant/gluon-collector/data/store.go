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

type Nodeinfostore interface {
	GetNodeinfo(nodeId string) (NodeInfo, error)
	GetStatistics(nodeId string) (StatisticsStruct, error)
	GetNodeinfos() []NodeInfo
	GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error)
	GetNodeNeighbours(nodeId string) (NeighbourStruct, error)
	LoadNodesFromFile(path string) error
	UpdateNodesJson()
}

type SimpleInMemoryStore struct {
	Nodeinfos       map[string]NodeInfo
	Statistics      map[string]StatisticsStruct
	StatusInfo      map[string]NodeStatusInfo
	NodesJsonPath   string
	CachedNodesJson string
	GatewayList     map[string]bool
	NeighbourInfos  map[string]NeighbourStruct
}

func NewSimpleInMemoryStore() *SimpleInMemoryStore {
	return &SimpleInMemoryStore{
		Nodeinfos:      make(map[string]NodeInfo),
		Statistics:     make(map[string]StatisticsStruct),
		StatusInfo:     make(map[string]NodeStatusInfo),
		NeighbourInfos: make(map[string]NeighbourStruct),
		GatewayList:    make(map[string]bool),
	}
}

func (s *SimpleInMemoryStore) LoadNodesFromFile(path string) error {
	s.NodesJsonPath = path
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
		Nodeinfos := nodeJsonInfo.Nodeinfo
		nodeStats := nodeJsonInfo.Statistics
		nodeStatus := NodeStatusInfo{
			Firstseen: nodeJsonInfo.Firstseen,
			Lastseen:  nodeJsonInfo.Lastseen,
			Online:    nodeJsonInfo.Flags.Online,
			Gateway:   nodeJsonInfo.Flags.Gateway,
		}
		s.Nodeinfos[nodeId] = Nodeinfos
		s.Statistics[nodeId] = nodeStats
		s.StatusInfo[nodeId] = nodeStatus
	}
	return nil
}

func (s *SimpleInMemoryStore) updateNodeStatusInfo(response ParsedResponse) {
	info, exists := s.StatusInfo[response.NodeId()]
	now := time.Now().Format(TimeFormat)
	_, isGw := s.GatewayList[response.NodeId()]
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
	s.StatusInfo[response.NodeId()] = info
}

func (s *SimpleInMemoryStore) GetNodeStatusInfo(nodeId string) (status NodeStatusInfo, err error) {
	stat, exists := s.StatusInfo[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s has no status info", nodeId)
	}
	status = stat
	return
}

func (s *SimpleInMemoryStore) GetNodeNeighbours(nodeId string) (neighbours NeighbourStruct, err error) {
	neighbourInfo, exists := s.NeighbourInfos[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s has no neighbour info", nodeId)
	}
	neighbours = neighbourInfo
	return
}

func (s *SimpleInMemoryStore) GetNodeinfo(nodeId string) (info NodeInfo, err error) {
	nodeinfo, exists := s.Nodeinfos[nodeId]
	info = nodeinfo
	if !exists {
		err = fmt.Errorf("NodeId %s does not exist", nodeId)
		return
	}
	return
}

func (s *SimpleInMemoryStore) GetNodeinfos() []NodeInfo {
	Nodeinfos := make([]NodeInfo, len(s.Nodeinfos))
	counter := 0
	for _, nodeinfo := range s.Nodeinfos {
		Nodeinfos[counter] = nodeinfo
		counter++
	}
	return Nodeinfos
}

func (s *SimpleInMemoryStore) GetStatistics(nodeId string) (Statistics StatisticsStruct, err error) {
	stats, exists := s.Statistics[nodeId]
	Statistics = stats
	if !exists {
		err = fmt.Errorf("NodeId %s has no Statistics", nodeId)
	}
	return
}

func (s *SimpleInMemoryStore) Routes() []httpserver.Route {
	var memoryStoreRoutes = []httpserver.Route{
		httpserver.Route{"NodeInfo", "GET", "/Nodeinfos/{nodeid}", s.GetNodeInfoRest},
		httpserver.Route{"Nodeinfos", "GET", "/Nodeinfos", s.GetNodeinfosRest},
		httpserver.Route{"NodeStatistics", "GET", "/Statistics/{nodeid}", s.GetNodeStatisticsRest},
		httpserver.Route{"NodesJson", "GET", "/nodes.json", s.GetNodesJsonRest},
		httpserver.Route{"NodesNeighbours", "GET", "/neighbours/{nodeid}", s.GetNodeNeighboursRest},
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

func (s *SimpleInMemoryStore) GetNodeNeighboursRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	neighbours, err := s.GetNodeNeighbours(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(neighbours)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}

func (s *SimpleInMemoryStore) GetNodeinfosRest(w http.ResponseWriter, r *http.Request) {
	Nodeinfos := s.GetNodeinfos()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Nodeinfos)
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
	w.Write([]byte(s.CachedNodesJson))
}

func (s *SimpleInMemoryStore) GetNodesJson() NodesJson {
	timestamp := time.Now().Format(TimeFormat)
	nodes := make(map[string]NodesJsonNode)
	for nodeId, nodeInfo := range s.Nodeinfos {
		stats := s.Statistics[nodeId]
		status := s.StatusInfo[nodeId]
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
			"path":  s.NodesJsonPath,
		}).Error("Error encoding json")
		return
	}
	s.CachedNodesJson = string(data)
	if s.NodesJsonPath == "" {
		return
	}
	err = ioutil.WriteFile(s.NodesJsonPath, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  s.NodesJsonPath,
		}).Error("Error writing nodes.json")
	}
}
