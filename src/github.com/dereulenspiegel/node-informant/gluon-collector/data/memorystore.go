package data

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/gorilla/mux"
)

type SimpleInMemoryStore struct {
	Nodeinfos       map[string]NodeInfo
	Statistics      map[string]*StatisticsStruct
	StatusInfo      map[string]NodeStatusInfo
	NodesJsonPath   string
	CachedNodesJson string
	GatewayList     map[string]bool
	NeighbourInfos  map[string]*NeighbourStruct
}

func NewSimpleInMemoryStore() *SimpleInMemoryStore {
	return &SimpleInMemoryStore{
		Nodeinfos:      make(map[string]NodeInfo),
		Statistics:     make(map[string]*StatisticsStruct),
		StatusInfo:     make(map[string]NodeStatusInfo),
		NeighbourInfos: make(map[string]*NeighbourStruct),
		GatewayList:    make(map[string]bool),
	}
}

func (s *SimpleInMemoryStore) GetNodeStatusInfo(nodeId string) (status NodeStatusInfo, err error) {
	stat, exists := s.StatusInfo[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s has no status info", nodeId)
	}
	status = stat
	return
}

func (s *SimpleInMemoryStore) GetNodeStatusInfos() []NodeStatusInfo {
	list := make([]NodeStatusInfo, 0, len(s.StatusInfo))
	for _, status := range s.StatusInfo {
		list = append(list, status)
	}
	return list
}

func (s *SimpleInMemoryStore) PutNodeStatusInfo(nodeId string, info NodeStatusInfo) {
	s.StatusInfo[nodeId] = info
}

func (s *SimpleInMemoryStore) GetStatistics(nodeId string) (Statistics StatisticsStruct, err error) {
	stats, exists := s.Statistics[nodeId]
	Statistics = *stats
	if !exists {
		err = fmt.Errorf("NodeId %s has no Statistics", nodeId)
	}
	return
}

func (s *SimpleInMemoryStore) GetAllStatistics() []StatisticsStruct {
	list := make([]StatisticsStruct, 0, len(s.Statistics))
	for _, statistics := range s.Statistics {
		list = append(list, *statistics)
	}
	return list
}

func (s *SimpleInMemoryStore) PutStatistics(statistics StatisticsStruct) {
	s.Statistics[statistics.NodeId] = &statistics
}

func (s *SimpleInMemoryStore) GetNodeNeighbours(nodeId string) (neighbours NeighbourStruct, err error) {
	neighbourInfo, exists := s.NeighbourInfos[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s has no neighbour info", nodeId)
	}
	neighbours = *neighbourInfo
	return
}

func (s *SimpleInMemoryStore) GetAllNeighbours() []NeighbourStruct {
	list := make([]NeighbourStruct, 0, len(s.NeighbourInfos))
	for _, neighbour := range s.NeighbourInfos {
		list = append(list, *neighbour)
	}
	return list
}

func (s *SimpleInMemoryStore) PutNodeNeighbours(neighbours NeighbourStruct) {
	s.NeighbourInfos[neighbours.NodeId] = &neighbours
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

func (s *SimpleInMemoryStore) PutNodeInfo(nodeInfo NodeInfo) {
	s.Nodeinfos[nodeInfo.NodeId] = nodeInfo
}

func (s *SimpleInMemoryStore) GetNodeinfos() []NodeInfo {
	list := make([]NodeInfo, 0, len(s.Nodeinfos))
	for _, nodeinfo := range s.Nodeinfos {
		list = append(list, nodeinfo)
	}
	return list
}

func (s *SimpleInMemoryStore) Routes() []httpserver.Route {
	var memoryStoreRoutes = []httpserver.Route{
		httpserver.Route{"NodeInfo", "GET", "/nodeinfos/{nodeid}", s.GetNodeInfoRest},
		httpserver.Route{"Nodeinfos", "GET", "/nodeinfos", s.GetNodeinfosRest},
		httpserver.Route{"NodeStatistics", "GET", "/Statistics/{nodeid}", s.GetNodeStatisticsRest},
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
