package data

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

type BoltStore struct {
	db     *bolt.DB
	bucket *bolt.Bucket
}

const NodeinfoBucket string = "nodeinfos"
const StatisticsBucket string = "statistics"

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Errorf("Error opening db")
		return nil, err
	}
	store := &BoltStore{db: db}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(BucketName))
		if err != nil {
			return err
		}
		store.bucket = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (b *BoltStore) Close() {
	b.db.Close()
}

func (b *BoltStore) put(key, bucket string, data interface{}) {
	bytes, err := json.Marshal(data)
	if err == nil {
		err = b.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			err := b.Put([]byte(key), bytes)
		})
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"key":    key,
			"bucket": bucket,
			"data":   data,
		}).Error("Error putting data into bolt store")
	}
}

func (b *BoltStore) get(key, bucket string, object interface{}) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(key))
		if v != nil {
			return fmt.Errorf("Can't find object for key %s", key)
		}
		err := json.Unmarshal(v, object)
		return err
	})
	return err
}

func (b *BoltStore) GetNodeInfo(nodeId string) (NodeInfo, error) {
	info := &NodeInfo{}
	err := b.get(nodeId, NodeinfoBucket, info)
	return *info, err
}

func (b *BoltStore) PutNodeInfo(nodeInfo NodeInfo) {
	b.put(nodeInfo.NodeId, NodeinfoBucket, nodeInfo)
}

func (b *BoltStore) allValues(bucket string, iterFunc func(string, []byte)) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			iterFunc(string(k), v)
		}

		return nil
	})
	return err
}

func (b *BoltStore) GetNodeInfos() []NodeInfo {
	allNodeinfos := make([]NodeInfo, 0, 500)
	err := b.allValues(NodeinfoBucket, func(key string, data []byte) {
		nodeinfo := &NodeInfo{}
		err := json.Unmarshal(data, nodeinfo)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"jsonString": string(data),
				"nodeid":     key,
			}).Error("Error unmarshalling json data")
		} else {
			allNodeinfos = append(allNodeinfos, nodeinfo)
		}
	})
	return allNodeinfos
}

func (b *BoltStore) GetStatistics(nodeId string) (StatisticsStruct, error) {
	statistics := &StatisticsStruct{}
	err := b.get(nodeId, StatisticsBucket, statistics)
	return *statistics, err
}

func (b *BoltStore) PutStatistics(statistics StatisticsStruct) {
	b.put(statistics.NodeId, StatisticsBucket, statistics)
}

func (b *BoltStore) GetAllStatistics() []StatisticsStruct {
	allStatistics := make([]StatisticsStruct, 0, 500)
	err := b.allValues(StatisticsBucket, func(key string, data []byte) {
		statistics := &StatisticsStruct{}
		err := json.Unmarshal(data, statistics)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"jsonString": string(data),
				"nodeid":     key,
			}).Error("Error unmarshalling json data")
		} else {
			allStatistics = append(allStatistics, statistics)
		}
	})
	return allStatistics
}
