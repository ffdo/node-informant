package meshviewer

import (
	"encoding/json"
	"os"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
)

type DataLoader struct {
	Store         *data.SimpleInMemoryStore
	NodesJsonPath string
}

func convertFromMeshviewerStatistics(nodeId string, in StatisticsStruct) data.StatisticsStruct {
	clients := data.ClientStatistics{
		Wifi:  in.Clients,
		Total: in.Clients,
	}
	memory := data.MemoryStatistics{
		Total: 100,
		Free:  uint64((float64(100) * in.MemoryUsage)),
	}
	return data.StatisticsStruct{
		NodeId:      nodeId,
		Traffic:     *in.Traffic,
		Clients:     clients,
		LoadAverage: in.Loadavg,
		RootFsUsage: in.RootfsUsage,
		Memory:      memory,
	}
}

func (l *DataLoader) LoadNodesFromFile(path string) error {
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
		Nodeinfos := nodeJsonInfo.Nodeinfo
		nodeStats := nodeJsonInfo.Statistics
		nodeStatus := data.NodeStatusInfo{
			Firstseen: nodeJsonInfo.Firstseen,
			Lastseen:  nodeJsonInfo.Lastseen,
			Online:    nodeJsonInfo.Flags.Online,
			Gateway:   nodeJsonInfo.Flags.Gateway,
		}
		l.Store.Nodeinfos[nodeId] = Nodeinfos
		l.Store.Statistics[nodeId] = convertFromMeshviewerStatistics(nodeId, nodeStats)
		l.Store.StatusInfo[nodeId] = nodeStatus
	}
	return nil
}
