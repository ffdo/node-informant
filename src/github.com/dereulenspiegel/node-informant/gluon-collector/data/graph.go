package data

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

type GraphGenerator struct {
	Store            *SimpleInMemoryStore
	cachedJsonString string
}

func FindInLinks(links []*GraphLink, sourceIndex, targetIndex int) (link *GraphLink, err error) {
	for _, item := range links {
		if item.Source == sourceIndex && item.Target == targetIndex {
			link = item
			err = nil
			return
		}
	}
	err = fmt.Errorf("Link not found")
	return
}

func (g *GraphGenerator) GenerateGraphJson() GraphJson {
	nodeTable := make(map[string]*GraphNode)

	y := 0
	for nodeId, neighbourInfo := range g.Store.NeighbourInfos {
		for mac, _ := range neighbourInfo.Batdv {
			nodeTable[mac] = &GraphNode{
				Id:     mac,
				NodeId: nodeId,
			}
			y = y + 1
		}
	}

	nodeList := make([]*GraphNode, 0, len(nodeTable))
	i := 0
	for _, item := range nodeTable {
		item.tableId = i
		nodeList = append(nodeList, item)
		i = i + 1
	}
	allLinks := make([]*GraphLink, 0, len(g.Store.NeighbourInfos)*5)

	for _, neighbours := range g.Store.NeighbourInfos {
		for ownMac, neighbour := range neighbours.Batdv {
			for peerMac, linkInfo := range neighbour.Neighbours {
				source, sourceExists := nodeTable[ownMac]
				target, targetExists := nodeTable[peerMac]
				if !sourceExists || !targetExists {
					log.WithFields(log.Fields{
						"source-mac": ownMac,
						"target-mac": peerMac,
					}).Debug("Tried to build link to unknown peer")
					continue
				}
				link := &GraphLink{
					Source: source.tableId,
					Target: target.tableId,
					Tq:     float64(linkInfo.Tq),
				}
				allLinks = append(allLinks, link)
			}
		}
	}

	bidirectionalLinks := make([]*GraphLink, 0, len(g.Store.NeighbourInfos)*5)
	unidirectionalLinks := make([]*GraphLink, 0, len(g.Store.NeighbourInfos))
	for _, link := range allLinks {
		_, err := FindInLinks(allLinks, link.Target, link.Source)
		if err != nil {
			link.Bidirect = false
			unidirectionalLinks = append(unidirectionalLinks, link)
		} else {
			link.Bidirect = true
			_, err := FindInLinks(allLinks, link.Source, link.Target)
			if err != nil {
				bidirectionalLinks = append(bidirectionalLinks, link)
			}
		}
	}

	allLinks = make([]*GraphLink, 0, len(bidirectionalLinks)+len(unidirectionalLinks))
	allLinks = append(allLinks, bidirectionalLinks...)
	allLinks = append(allLinks, unidirectionalLinks...)

	for _, link := range allLinks {
		if link == nil {
			log.Warnf("Link is nil!")
			continue
		}
		source := nodeList[link.Source]
		target := nodeList[link.Target]
		_, sourceGW := g.Store.GatewayList[source.Id]
		_, targetGW := g.Store.GatewayList[target.Id]
		if sourceGW || targetGW {
			link.Vpn = true
		}
	}

	batGraph := BatadvGraph{
		Multigraph: false,
		Directed:   false,
		Links:      allLinks,
		Nodes:      nodeList,
	}

	graphJson := GraphJson{
		Batadv:  batGraph,
		Version: 1,
	}
	return graphJson
}

func (g *GraphGenerator) UpdateGraphJson() {
	graph := g.GenerateGraphJson()
	jsonBytes, err := json.Marshal(graph)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to marshal graph.json")
		return
	}
	g.cachedJsonString = string(jsonBytes)
}

func (g *GraphGenerator) GetGraphJsonRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(g.cachedJsonString))
}

func (g *GraphGenerator) Routes() []httpserver.Route {
	var graphRoutes = []httpserver.Route{
		httpserver.Route{"GraphJson", "GET", "/graph.json", g.GetGraphJsonRest},
	}
	return graphRoutes
}
