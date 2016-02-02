package store

import (
	"fmt"
	"time"

	"github.com/ffdo/node-informant/sensor/data"
	"github.com/ffdo/node-informant/sensor/process"
	"github.com/olebedev/config"
)

var (
	DB              Storage
	storageCreators map[string]StorageCreator
)

func init() {
	storageCreators = make(map[string]StorageCreator)
}

func RegisterStorageEngine(name string, creator StorageCreator) {
	storageCreators[name] = creator
}

type NodeID string

type Storage interface {
	UpdateNodeData(nodeID NodeID, value data.NodeData)
	ExpireNodeData(duration time.Duration) []NodeID
	GetNodeData(nodeID NodeID) (data.NodeData, error)
	GetAllNodeData() []data.NodeData
}

type StorageCreator func(*config.Config) (Storage, error)

func StoreReceivedData(packet data.NodeData) error {
	nodeId := packet.NodeId()
	if nodeId == "" {
		return fmt.Errorf("NodeId was empty")
	}
	DB.UpdateNodeData(NodeID(nodeId), packet)
	return nil
}

func ConfigureStorage(cfg *config.Config) error {
	name, err := cfg.String("type")
	if err != nil {
		return err
	}
	if creator, exists := storageCreators[name]; exists {
		if DB, err = creator(cfg); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Storage engine %s is unknwon", name)
	}
	process.RegisterProcessFunction(StoreReceivedData)
	return nil
}
