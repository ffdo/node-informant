package data

import (
	"encoding/json"
	"os"
	"path"
)

const NodeinfoFilename string = "Nodeinfo.json"
const StatisticsFilename string = "Statistics.json"
const NeighboursFilename string = "Neighbours.json"
const NodeStatusFilename string = "Status.json"

type JsonPersister struct {
	persistFolder string
}

func NewJsonPersister(path string) (*JsonPersister, error) {
	return &JsonPersister{persistFolder: path}, nil
}

func ensureFolder(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func openFile(path string) (*os.File, error) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		return file, err
	} else if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	return file, err
}

func (j *JsonPersister) persistData(fileName string, data interface{}) error {
	ensureFolder(j.persistFolder)
	file, err := os.Open(path.Join(j.persistFolder, fileName))
	if err != nil {
		return err
	}
	return json.NewEncoder(file).Encode(data)
}

func (j *JsonPersister) Perists(db Nodeinfostore) error {
	ensureFolder(j.persistFolder)
	if err := j.persistData(NodeinfoFilename, db.GetNodeInfos()); err != nil {
		return err
	}
	if err := j.persistData(StatisticsFilename, db.GetAllStatistics()); err != nil {
		return err
	}
	if err := j.persistData(NeighboursFilename, db.GetAllNeighbours()); err != nil {
		return err
	}
	if err := j.persistData(NodeStatusFilename, db.GetNodeStatusInfos()); err != nil {
		return err
	}
	return nil
}

func (j *JsonPersister) loadData(fileName string, data interface{}) error {
	file, err := os.Open(path.Join(j.persistFolder, fileName))
	if err != nil {
		return err
	}
	return json.NewDecoder(file).Decode(data)
}

func (j *JsonPersister) Load(db Nodeinfostore) error {
	allNodeinfos := make([]NodeInfo, 0, 500)
	allStatistics := make([]StatisticsStruct, 0, 500)
	allNeighbours := make([]NeighbourStruct, 0, 500)
	allStatus := make([]NodeStatusInfo, 0, 500)

	if err := j.loadData(NodeinfoFilename, &allNodeinfos); err != nil {
		return err
	} else {
		for _, info := range allNodeinfos {
			db.PutNodeInfo(info)
		}
	}
	if err := j.persistData(StatisticsFilename, &allStatistics); err != nil {
		return err
	} else {
		for _, statistic := range allStatistics {
			db.PutStatistics(statistic)
		}
	}
	if err := j.persistData(NeighboursFilename, &allNeighbours); err != nil {
		return err
	} else {
		for _, neighbour := range allNeighbours {
			db.PutNodeNeighbours(neighbour)
		}
	}
	if err := j.persistData(NodeStatusFilename, &allStatus); err != nil {
		return err
	} else {
		for _, status := range allStatus {
			db.PutNodeStatusInfo(status.NodeId, status)
		}
	}
	return nil
}
