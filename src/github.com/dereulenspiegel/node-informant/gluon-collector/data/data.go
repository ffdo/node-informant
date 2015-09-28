package data

type ParsedResponse interface {
	Type() string
	ParsedData() interface{}
	NodeId() string
}

type NodeinfoResponse struct {
	Nodeinfo NodeInfo
}

func (n NodeinfoResponse) Type() string {
	return "nodeinfo"
}

func (n NodeinfoResponse) ParsedData() interface{} {
	return n.Nodeinfo
}

func (n NodeinfoResponse) NodeId() string {
	return n.Nodeinfo.NodeId
}

type StatisticsResponse struct {
	Statistics StatisticsStruct
}

func (s StatisticsResponse) Type() string {
	return "statistics"
}

func (s StatisticsResponse) ParsedData() interface{} {
	return s.Statistics
}

func (s StatisticsResponse) NodeId() string {
	return s.Statistics.NodeId
}

type NodeFlags struct {
	Gateway bool `json:"gateway"`
	Online  bool `json:"online"`
}
type NodesJsonNode struct {
	Nodeinfo   NodeInfo         `json:"nodeinfo"`
	Statistics StatisticsStruct `json:"statistics"`
	Flags      NodeFlags        `json:"flags"`
	Lastseen   string           `json:"lastseen"`
	Firstseen  string           `json:"firstseen"`
}

type NodesJson struct {
	Timestamp string                   `json:"timestamp"`
	Version   int                      `json:"version"`
	Nodes     map[string]NodesJsonNode `json:"nodes"`
}
