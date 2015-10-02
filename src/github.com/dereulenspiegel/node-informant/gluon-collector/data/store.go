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
	GetStatistics(nodeId string) (StatisticsStruct, error)
	GetNodeinfos() []NodeInfo
	GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error)
	GetNodeNeighbours(nodeId string) (NeighbourStruct, error)
	LoadNodesFromFile(path string) error
	UpdateNodesJson()
}
