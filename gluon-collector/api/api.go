package api

import (
	"encoding/json"
	"net/http"

	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/httpserver"
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
		httpserver.Route{"AllStatistics", "GET", "/statistics", h.GetAllStatistics},
		httpserver.Route{"AllNodeStatus", "GET", "/nodestatus", h.GetAllNodeStatus},
		httpserver.Route{"NodeStatus", "GET", "/nodestatus/{nodeid}", h.GetNodeStatus},
	}
	return apiRoutes
}

func respond(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondOK(w http.ResponseWriter, data interface{}) {
	respond(w, data, http.StatusOK)
}

func respondMissing(w http.ResponseWriter, data error) {
	respond(w, data, http.StatusNotFound)
}

func (h *HttpApi) GetAllNodeStatus(w http.ResponseWriter, r *http.Request) {
	respondOK(w, h.Store.GetNodeStatusInfos())
}

func (h *HttpApi) GetNodeStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status, err := h.Store.GetNodeStatusInfo(vars["nodeid"])
	if err == nil {
		respondOK(w, status)
	} else {
		respondMissing(w, err)
	}
}

func (h *HttpApi) GetAllStatistics(w http.ResponseWriter, r *http.Request) {
	respondOK(w, h.Store.GetAllStatistics())
}

func (h *HttpApi) GetAllNeighboursRest(w http.ResponseWriter, r *http.Request) {
	respondOK(w, h.Store.GetAllNeighbours())
}

func (h *HttpApi) GetNodeStatisticsRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stats, err := h.Store.GetStatistics(vars["nodeid"])
	if err == nil {
		respondOK(w, stats)
	} else {
		respondMissing(w, err)
	}
}

func (n *HttpApi) GetNodeNeighboursRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	neighbours, err := n.Store.GetNodeNeighbours(vars["nodeid"])
	if err == nil {
		respondOK(w, neighbours)
	} else {
		respondMissing(w, err)
	}
}

func (n *HttpApi) GetNodeinfosRest(w http.ResponseWriter, r *http.Request) {
	respondOK(w, n.Store.GetNodeInfos())
}

func (n *HttpApi) GetNodeInfoRest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nodeinfo, err := n.Store.GetNodeInfo(vars["nodeid"])
	if err == nil {
		respondOK(w, nodeinfo)
	} else {
		respondMissing(w, err)
	}
}
