package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/gorilla/mux"
	"github.com/muesli/cache2go"
)

type SimpleInMemoryStore struct {
	Nodeinfos map[string]NodeInfo
	//Statistics      map[string]*StatisticsStruct
	statistics      *cache2go.CacheTable
	StatusInfo      map[string]NodeStatusInfo
	NodesJsonPath   string
	CachedNodesJson string
	GatewayList     map[string]bool
	neighbourCache  *cache2go.CacheTable
	//NeighbourInfos  map[string]*NeighbourStruct
}

func NewSimpleInMemoryStore() *SimpleInMemoryStore {
	return &SimpleInMemoryStore{
		Nodeinfos: make(map[string]NodeInfo),
		//Statistics: make(map[string]*StatisticsStruct),
		statistics: cache2go.Cache("statistics"),
		StatusInfo: make(map[string]NodeStatusInfo),
		//NeighbourInfos: make(map[string]*NeighbourStruct),
		neighbourCache: cache2go.Cache("neighbours"),
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
	data, err := s.statistics.Value(nodeId)
	if err != nil {
		err = fmt.Errorf("NodeId %s has no Statistics", nodeId)
		return
	}
	Statistics = *data.Data().(*StatisticsStruct)
	return
}

func (s *SimpleInMemoryStore) GetAllStatistics() []StatisticsStruct {
	list := make([]StatisticsStruct, 0, s.statistics.Count())
	s.neighbourCache.Foreach(func(key interface{}, item *cache2go.CacheItem) {
		list = append(list, *item.Data().(*StatisticsStruct))
	})
	return list
}

func (s *SimpleInMemoryStore) PutStatistics(statistics StatisticsStruct) {
	s.statistics.Add(statistics.NodeId,
		time.Second*time.Duration(conf.UInt("announced.interval.statistics", 300)*2),
		&statistics)
	//s.Statistics[statistics.NodeId] = &statistics
}

func (s *SimpleInMemoryStore) GetNodeNeighbours(nodeId string) (neighbours NeighbourStruct, err error) {
	data, err := s.neighbourCache.Value(nodeId)
	if err != nil {
		err = fmt.Errorf("NodeId %s has no neighbour info", nodeId)
		return
	}
	neighbours = *data.Data().(*NeighbourStruct)
	return
}

func (s *SimpleInMemoryStore) GetAllNeighbours() []NeighbourStruct {
	list := make([]NeighbourStruct, 0, s.neighbourCache.Count())
	s.neighbourCache.Foreach(func(key interface{}, item *cache2go.CacheItem) {
		list = append(list, *item.Data().(*NeighbourStruct))
	})
	return list
	/*for _, neighbour := range s.NeighbourInfos {
		list = append(list, *neighbour)
	}
	return list*/
}

func (s *SimpleInMemoryStore) PutNodeNeighbours(neighbours NeighbourStruct) {
	s.neighbourCache.Add(neighbours.NodeId,
		time.Second*time.Duration(conf.UInt("announced.interval.statistics", 300)*2),
		&neighbours)
	//s.NeighbourInfos[neighbours.NodeId] = &neighbours
}

func (s *SimpleInMemoryStore) GetNodeInfo(nodeId string) (info NodeInfo, err error) {
	nodeinfo, exists := s.Nodeinfos[nodeId]
	if !exists {
		err = fmt.Errorf("NodeId %s does not exist", nodeId)
		return
	}
	info = nodeinfo
	return
}

func (s *SimpleInMemoryStore) PutNodeInfo(nodeInfo NodeInfo) {
	s.Nodeinfos[nodeInfo.NodeId] = nodeInfo
}

func (s *SimpleInMemoryStore) GetNodeInfos() []NodeInfo {
	list := make([]NodeInfo, 0, len(s.Nodeinfos))
	for _, nodeinfo := range s.Nodeinfos {
		list = append(list, nodeinfo)
	}
	return list
}

func (s *SimpleInMemoryStore) PutGateway(mac string) {
	s.GatewayList[mac] = true
}

func (s *SimpleInMemoryStore) IsGateway(mac string) bool {
	isGateway, exists := s.GatewayList[mac]
	return exists && isGateway
}

func (s *SimpleInMemoryStore) RemoveGateway(mac string) {
	delete(s.GatewayList, mac)
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
	Nodeinfos := s.GetNodeInfos()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Nodeinfos)
}

func (s *SimpleInMemoryStore) GetNodeInfoRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	nodeinfo, err := s.GetNodeInfo(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nodeinfo)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}
