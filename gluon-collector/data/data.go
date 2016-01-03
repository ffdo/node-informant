package data

type ParsedResponse interface {
	ParsedData() interface{}
	NodeId() string
}

type ParsedResponseReader func(store Nodeinfostore, response ParsedResponse)

type NodeinfoResponse struct {
	Nodeinfo NodeInfo
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

func (s StatisticsResponse) ParsedData() interface{} {
	return s.Statistics
}

func (s StatisticsResponse) NodeId() string {
	return s.Statistics.NodeId
}

type NeighbourReponse struct {
	Neighbours *NeighbourStruct
}

func (n NeighbourReponse) ParsedData() interface{} {
	return n.Neighbours
}

func (n NeighbourReponse) NodeId() string {
	return n.Neighbours.NodeId
}

type ErroredResponse struct{}

func (n ErroredResponse) ParsedData() interface{} {
	return nil
}

func (n ErroredResponse) NodeId() string {
	return ""
}
