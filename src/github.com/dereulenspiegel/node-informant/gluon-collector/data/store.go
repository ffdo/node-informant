package data

const TimeFormat string = "2006-01-02T15:04:05"

type NodeStatusInfo struct {
	Firstseen string
	Lastseen  string
	Online    bool
	Gateway   bool
}

type Nodeinfostore interface {
	GetNodeInfo(nodeId string) (NodeInfo, error)
	GetNodeInfos() []NodeInfo
	PutNodeInfo(nodeInfo NodeInfo)

	GetStatistics(nodeId string) (StatisticsStruct, error)
	GetAllStatistics() []StatisticsStruct
	PutStatistics(statistics StatisticsStruct)

	GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error)
	GetNodeStatusInfos() []NodeStatusInfo
	PutNodeStatusInfo(nodeId string, info NodeStatusInfo)

	GetNodeNeighbours(nodeId string) (NeighbourStruct, error)
	GetAllNeighbours() []NeighbourStruct
	PutNodeNeighbours(neighbours NeighbourStruct)

	PutGateway(mac string)
	IsGateway(mac string) bool
	RemoveGateway(mac string)
}

type NodeFileImporter interface {
	LoadNodesFromFile(path string) error
}
