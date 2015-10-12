package data

import "time"

const LegacyTimeFormat string = "2006-01-02T15:04:05"
const TimeFormat string = time.RFC3339

type NodeStatusInfo struct {
	Firstseen string
	Lastseen  string
	Online    bool
	Gateway   bool
	NodeId    string
}

// Nodeinfostore needs to implemented by all types which want to store node
// information. Currently only one Nodeinfostore is used in the whole application
// at a time.
type Nodeinfostore interface {

	// GetNodeInfo retrieves the rather static NodeInfos for a single node specified
	// by the node id. If there is no information for this node id an error is
	// returned.
	GetNodeInfo(nodeId string) (NodeInfo, error)

	// GetNodeInfos retrieves all stored NodeInfos or an empty slice if no nodeInfo
	// is available.
	GetNodeInfos() []NodeInfo

	// PutNodeInfo stores a NodeInfo object in the data store. The NodeId is retrieved
	// from the NodeInfo object itself.
	PutNodeInfo(nodeInfo NodeInfo)

	// GetStatistics retrieves the latest received statistics for the specified nodeId
	// or returns an error of no statistics are available for this node id.
	GetStatistics(nodeId string) (StatisticsStruct, error)

	// GetAllStatistics returns all stored node statistics or an empty slice if no
	// statistics are available.
	GetAllStatistics() []StatisticsStruct

	// PutStatistics store the Statistics object in the data store. The node id is
	// retrieved from the StatisticsStruct.
	PutStatistics(statistics StatisticsStruct)

	// GetNodeStatusInfo retrieves the current status information for the node id
	// or an error if no status information is available.
	GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error)

	// GetNodeStatusInfos retrieves the stored NodeStatusInfo for all nodes or an
	// empty slice if no NodeStatusInfo is available.
	GetNodeStatusInfos() []NodeStatusInfo

	// PutNodeStatusInfo stores the NodeStatusInfo under the specified node id.
	// Note that all NodeStatusInfo objects need to have the NodeId set the same
	// node id. This is not checked or handled currently.
	PutNodeStatusInfo(nodeId string, info NodeStatusInfo)

	// GetNodeNeighbours retrives the mesh neighbour information for the specified
	// node id or returns an error if no mesh neighbour information is available
	// for the specified node id.
	GetNodeNeighbours(nodeId string) (NeighbourStruct, error)

	// GetAllNeighbours retrieves all stored mesh neighbour information or an empty
	// slice if no mesh neighbour information are available.
	GetAllNeighbours() []NeighbourStruct

	// PutNodeNeighbours stores the mesh neighbour information for the node id specified
	// in the NeighbourStruct.
	PutNodeNeighbours(neighbours NeighbourStruct)

	// PutGateway stores the information that the mac addressed represented by the
	// given string is gateway.
	PutGateway(mac string)

	// IsGateway checks if the the given mac addressed represented by the string is
	// a getway.
	IsGateway(mac string) bool

	// RemoveGateway removes the given mac address as a gateway, indicating that the
	// node with this mac address is not a gateway any more.
	RemoveGateway(mac string)

	NotifyNodeOffline(handler func(string))
}

// NodeFileImporter is a poor name for this. This interface can be implemented by
// all types which want to be able to import node information from legacy data sources
// like ffmap-backend into the current datastore.
type NodeFileImporter interface {
	LoadNodesFromFile(path string) error
}
