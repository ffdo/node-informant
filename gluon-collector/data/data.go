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
	Statistics *StatisticsStruct
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
	Neighbours *NeighbourStruct
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

type ErroredResponse struct{}

func (n ErroredResponse) Type() string {
	return "errored"
}

func (n ErroredResponse) ParsedData() interface{} {
	return nil
}

func (n ErroredResponse) NodeId() string {
	return ""
}
