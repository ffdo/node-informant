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

func convertFromMeshviewerStatistics(nodeId string, in *StatisticsStruct) data.StatisticsStruct {
	if in == nil {
		return data.StatisticsStruct{}
	}
	clients := data.ClientStatistics{
		Wifi:  in.Clients,
		Total: in.Clients,
	}
	memory := data.MemoryStatistics{
		Total: 100,
		Free:  uint64((float64(100) * in.MemoryUsage)),
	}
	statistics := data.StatisticsStruct{
		NodeId:      nodeId,
		Clients:     clients,
		LoadAverage: in.Loadavg,
		RootFsUsage: in.RootfsUsage,
		Memory:      memory,
	}
	if in.Traffic != nil {
		statistics.Traffic = *in.Traffic
	}
	return statistics
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

		if nodeStats != nil {
			statistics := convertFromMeshviewerStatistics(nodeId, nodeStats)

			l.Store.PutStatistics(statistics)
		}
		l.Store.PutNodeStatusInfo(nodeId, nodeStatus)
	}
	return nil
}
