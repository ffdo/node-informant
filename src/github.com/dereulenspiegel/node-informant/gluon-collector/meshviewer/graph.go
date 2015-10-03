package meshviewer

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

type GraphNode struct {
	Id      string `json:"id"`
	NodeId  string `json:"node_id"`
	tableId int
}

type GraphLink struct {
	Bidirect bool    `json:"bidirect"`
	Source   int     `json:"source"`
	Target   int     `json:"target"`
	Tq       float64 `json:"tq"`
	Vpn      bool    `json:"vpn"`
}

type BatadvGraph struct {
	Multigraph bool          `json:"multigraph"`
	Nodes      []*GraphNode  `json:"nodes"`
	Directed   bool          `json:"directed"`
	Links      []*GraphLink  `json:"links"`
	Graph      []interface{} `json:"graph"`
}

type GraphJson struct {
	Batadv  BatadvGraph `json:"batadv"`
	Version uint64      `json:"version"`
}

type GraphGenerator struct {
	Store            data.Nodeinfostore
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

func (g *GraphGenerator) buildNodeTableAndList() (map[string]*GraphNode, []*GraphNode) {
	allNeighbours := g.Store.GetAllNeighbours()
	nodeList := make([]*GraphNode, 0, len(allNeighbours))
	for _, neighbourInfo := range allNeighbours {
		for mac, _ := range neighbourInfo.Batadv {
			node := &GraphNode{
				Id:     mac,
				NodeId: neighbourInfo.NodeId,
			}
			nodeList = append(nodeList, node)
		}
	}

	nodeTable := make(map[string]*GraphNode)
	for i, node := range nodeList {
		node.tableId = i
		nodeTable[node.Id] = node
	}
	return nodeTable, nodeList
}

func calculateTq(tqSource, tqTarget int) float64 {
	min := math.Min(float64(tqSource), float64(tqTarget))
	return (1.0 / (min / 255.0))
}

func (g *GraphGenerator) buildLink(nodeTable map[string]*GraphNode, sourceMac, targetMac string, sourceLinkInfo data.BatmanLink) *GraphLink {
	sourceNode, sourceExists := nodeTable[sourceMac]
	targetNode, targetExists := nodeTable[targetMac]

	if !sourceExists {
		log.Warnf("Building link with nonexistant source node %s", sourceMac)
		return nil
	}
	if !targetExists {
		log.Warnf("Building link with nonexistant target node %s", targetMac)
		return nil
	}
	link := &GraphLink{
		Bidirect: false,
		Source:   sourceNode.tableId,
		Target:   targetNode.tableId,
		Vpn:      false,
		Tq:       float64(1.0 / (float64(sourceLinkInfo.Tq) / 255.0)),
	}
	targetNeighbourInfo, err := g.Store.GetNodeNeighbours(targetNode.NodeId)
	if err != nil {
		log.Warnf("Can't find neighbourinfos for nodeId %s", targetNode.NodeId)
	}

	targetBatInfo, exists := targetNeighbourInfo.Batadv[targetMac]
	if !exists {
		log.Warnf("Can't find Batadv links for %s", targetMac)
		link.Bidirect = false
	} else {
		targetLinkInfo, exists := targetBatInfo.Neighbours[sourceMac]
		if !exists {
			log.Warnf("Can't find linkinfo from %s to %s", targetMac, sourceMac)
			link.Bidirect = false
		} else {
			link.Bidirect = true
			// TODO How do we calculate a valid Tq value for meshviewer
			link.Tq = calculateTq(sourceLinkInfo.Tq, targetLinkInfo.Tq)
		}
	}
	return link
}

func (g *GraphGenerator) GenerateGraph() GraphJson {
	nodeTable, nodeList := g.buildNodeTableAndList()

	allNeighbours := g.Store.GetAllNeighbours()

	bidirectionalLinks := make([]*GraphLink, 0, len(allNeighbours))
	unidirectionalLinks := make([]*GraphLink, 0, len(allNeighbours))
	for _, neighbourInfo := range allNeighbours {
		for ownMac, batInfo := range neighbourInfo.Batadv {
			for peerMac, linkInfo := range batInfo.Neighbours {
				link := g.buildLink(nodeTable, ownMac, peerMac, linkInfo)
				if link == nil {
					log.Warnf("Couldn't form link between %s and %s", ownMac, peerMac)
				} else if link.Bidirect {
					bidirectionalLinks = append(bidirectionalLinks, link)
				} else {
					unidirectionalLinks = append(unidirectionalLinks, link)
				}
			}
		}
	}

	allLinks := make([]*GraphLink, 0, len(unidirectionalLinks)+len(bidirectionalLinks))
	allLinks = append(allLinks, bidirectionalLinks...)
	allLinks = append(allLinks, unidirectionalLinks...)
	batGraph := BatadvGraph{
		Multigraph: false,
		Directed:   false,
		Nodes:      nodeList,
		Links:      allLinks,
	}
	graph := GraphJson{
		Batadv:  batGraph,
		Version: 1,
	}
	return graph
}

func (g *GraphGenerator) UpdateGraphJson() {
	graph := g.GenerateGraph()
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
