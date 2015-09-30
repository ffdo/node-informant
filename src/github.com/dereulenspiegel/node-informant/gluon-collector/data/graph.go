package data

import "fmt"

type GraphGenerator struct {
	Store *SimpleInMemoryStore
}

func FindInLinks(links []GraphLink, sourceIndex, targetIndex int) (link GraphLink, err error) {
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

func (g *GraphGenerator) GenerateGraphJson() {
	nodeTable := make(map[string]GraphNode)
	i := 0
	for nodeId, _ := range g.Store.neighbourInfos {
		nodeinfo := g.Store.nodeinfos[nodeId]
		nodeTable[nodeinfo.Network.Mac] = GraphNode{
			Id:      nodeinfo.Network.Mac,
			NodeId:  nodeId,
			tableId: i,
		}
		i = i + 1
	}

	nodeList := make([]GraphNode, len(nodeTable))
	for _, item := range nodeTable {
		nodeList[item.tableId] = item
	}

	allLinks := make([]GraphLink, 0, len(g.Store.neighbourInfos)*5)

	for _, neighbours := range g.Store.neighbourInfos {
		for ownMac, neighbour := range neighbours.Batdv {
			for peerMac, linkInfo := range neighbour.Neighbours {
				link := GraphLink{
					Source: nodeTable[ownMac].tableId,
					Target: nodeTable[peerMac].tableId,
					Tq:     float64(linkInfo.Tq),
				}
				allLinks = append(allLinks, link)
			}
		}
	}

	bidirectionalLinks := make([]GraphLink, 0, len(g.Store.neighbourInfos)*5)
	unidirectionalLinks := make([]GraphLink, 0, len(g.Store.neighbourInfos))
	for _, link := range allLinks {
		_, err := FindInLinks(allLinks, link.Target, link.Source)
		if err != nil {
			unidirectionalLinks = append(unidirectionalLinks, link)
		} else {
			link.Bidirect = true
			bidirectionalLinks = append(bidirectionalLinks, link)
		}
	}

	allLinks = make([]GraphLink, len(bidirectionalLinks)+len(unidirectionalLinks))
	allLinks = append(allLinks, bidirectionalLinks...)
	allLinks = append(allLinks, unidirectionalLinks...)

	for _, link := range allLinks {
		source := nodeList[link.Source]
		target := nodeList[link.Target]
		_, sourceGW := g.Store.gatewayList[source.Id]
		_, targetGW := g.Store.gatewayList[target.Id]
		if sourceGW || targetGW {
			link.Vpn = true
		}
	}
}
