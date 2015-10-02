package meshviewer

import (
	"encoding/json"
	"os"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type FFMapBackendDataLoader struct {
	Store         data.Nodeinfostore
	NodesJsonPath string
}

func convertFromMeshviewerStatistics(nodeId string, in *StatisticsStruct) *data.StatisticsStruct {
	clients := data.ClientStatistics{
		Wifi:  in.Clients,
		Total: in.Clients,
	}
	memory := data.MemoryStatistics{
		Total: 100,
		Free:  uint64((float64(100) * in.MemoryUsage)),
	}
	return &data.StatisticsStruct{
		NodeId:      nodeId,
		Traffic:     *in.Traffic,
		Clients:     clients,
		LoadAverage: in.Loadavg,
		RootFsUsage: in.RootfsUsage,
		Memory:      memory,
	}
}

func (l *FFMapBackendDataLoader) LoadNodesFromFile(path string) error {
	l.NodesJsonPath = path
	nodesFile, err := os.Open(path)
	defer nodesFile.Close()
	if err != nil {
		return err
	}
	jsonParser := json.NewDecoder(nodesFile)
	nodesJson := &NodesJson{}
	if err = jsonParser.Decode(nodesJson); err != nil {
		return err
	}
	for nodeId, nodeJsonInfo := range nodesJson.Nodes {
		nodeinfos := nodeJsonInfo.Nodeinfo
		nodeStats := nodeJsonInfo.Statistics
		nodeStatus := data.NodeStatusInfo{
			Firstseen: nodeJsonInfo.Firstseen,
			Lastseen:  nodeJsonInfo.Lastseen,
			Online:    nodeJsonInfo.Flags.Online,
			Gateway:   nodeJsonInfo.Flags.Gateway,
		}
		l.Store.PutNodeInfo(nodeinfos)
		l.Store.PutStatistics(*convertFromMeshviewerStatistics(nodeId, nodeStats))
		l.Store.PutNodeStatusInfo(nodeId, nodeStatus)
	}
	return nil
}
