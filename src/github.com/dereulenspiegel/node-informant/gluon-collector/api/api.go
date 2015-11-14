package api

import (
	"encoding/json"
	"net/http"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/gorilla/mux"
)

type HttpApi struct {
	Store data.Nodeinfostore
}

func (h *HttpApi) Routes() []httpserver.Route {
	var apiRoutes = []httpserver.Route{
		httpserver.Route{"NodeInfo", "GET", "/nodeinfos/{nodeid}", h.GetNodeInfoRest},
		httpserver.Route{"Nodeinfos", "GET", "/nodeinfos", h.GetNodeinfosRest},
		httpserver.Route{"NodeStatistics", "GET", "/statistics/{nodeid}", h.GetNodeStatisticsRest},
		httpserver.Route{"NodesNeighbours", "GET", "/neighbours/{nodeid}", h.GetNodeNeighboursRest},
		httpserver.Route{"AllNeighbours", "GET", "/neighbours", h.GetAllNeighboursRest},
		httpserver.Route{"AllStatistucs", "GET", "/statistics", h.GetAllStatistics},
		httpserver.Route{"AllNodeStatus", "GET", "/nodestatus", h.GetAllNodeStatus},
		httpserver.Route{"NodeStatus", "GET", "/nodestatus/{nodeid}", h.GetNodeStatus},
	}
	return apiRoutes
}

func respondJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func respondMissing(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(err)
}

func (h *HttpApi) GetAllNodeStatus(w http.ResponseWriter, r *http.Request) {
	respondJson(w, h.Store.GetNodeStatusInfos())
}

func (h *HttpApi) GetNodeStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	status, err := h.Store.GetNodeStatusInfo(nodeid)
	if err == nil {
		respondJson(w, status)
	} else {
		respondMissing(w, err)
	}
}

func (h *HttpApi) GetAllStatistics(w http.ResponseWriter, r *http.Request) {
	respondJson(w, h.Store.GetAllStatistics())
}

func (h *HttpApi) GetAllNeighboursRest(w http.ResponseWriter, r *http.Request) {
	neighbours := h.Store.GetAllNeighbours()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(neighbours)
}

func (h *HttpApi) GetNodeStatisticsRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	stats, err := h.Store.GetStatistics(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}

func (n *HttpApi) GetNodeNeighboursRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	neighbours, err := n.Store.GetNodeNeighbours(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(neighbours)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}

func (n *HttpApi) GetNodeinfosRest(w http.ResponseWriter, r *http.Request) {
	Nodeinfos := n.Store.GetNodeInfos()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Nodeinfos)
}

func (n *HttpApi) GetNodeInfoRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeid := vars["nodeid"]
	nodeinfo, err := n.Store.GetNodeInfo(nodeid)
	if err == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(nodeinfo)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err)
	}
}
