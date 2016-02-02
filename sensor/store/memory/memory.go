package store

import (
	"sync"
	"time"

	"github.com/ffdo/node-informant/sensor/data"
	"github.com/ffdo/node-informant/sensor/store"
	"github.com/olebedev/config"
)

func init() {
	store.RegisterStorageEngine("memory", CreateInMemoryStore)
}

func CreateInMemoryStore(cfg *config.Config) (store.Storage, error) {
	return &InMemoryStore{
		store: make(map[store.NodeID]data.NodeData),
		lock:  &sync.Mutex{},
	}, nil
}

type InMemoryStore struct {
	store map[store.NodeID]data.NodeData
	lock  *sync.Mutex
}

func (i *InMemoryStore) UpdateNodeData(nodeID store.NodeID, value data.NodeData) {
	i.lock.Lock()
	if oldNodeData, exists := i.store[nodeID]; exists {
		oldNodeData.Merge(value)
	} else {
		i.store[nodeID] = value
	}
	i.lock.Unlock()
}

func (i *InMemoryStore) ExpireNodeData(duration time.Duration) []store.NodeID {
	return []store.NodeID{}
}

func (i *InMemoryStore) GetNodeData(nodeID store.NodeID) (data.NodeData, error) {
	return data.NodeData{}, nil
}

func (i *InMemoryStore) GetAllNodeData() []data.NodeData {
	i.lock.Lock()
	allNodes := make([]data.NodeData, 0, len(i.store))
	for _, v := range i.store {
		allNodes = append(allNodes, v)
	}
	i.lock.Unlock()
	return allNodes
}
