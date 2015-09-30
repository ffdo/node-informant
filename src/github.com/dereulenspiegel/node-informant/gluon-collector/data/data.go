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

type NeighbourReponse struct {
	Neighbours NeighbourStruct
}

func (n NeighbourReponse) Type() string {
	return "neighbours"
}

func (n NeighbourReponse) ParsedData() interface{} {
	return n.Neighbours
}

func (n NeighbourReponse) NodeId() string {
	return n.Neighbours.NodeId
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
	Multigraph bool        `json:"multigraph"`
	Nodes      []GraphNode `json:"nodes"`
	Directed   bool        `json:"directed"`
	Links      []GraphLink `json:"links"`
}

type GraphJson struct {
	Batadv BatadvGraph `json:"batadv"`
}
