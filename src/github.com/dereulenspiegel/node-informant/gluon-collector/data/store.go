package data

const TimeFormat string = "2006-01-02T15:04:05"

type NodeStatusInfo struct {
	Firstseen string
	Lastseen  string
	Online    bool
	Gateway   bool
}

type Nodeinfostore interface {
	GetNodeinfo(nodeId string) (NodeInfo, error)
	GetNodeinfos() []NodeInfo
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
}
