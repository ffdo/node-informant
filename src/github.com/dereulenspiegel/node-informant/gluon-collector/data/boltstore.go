package data

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

// BoltStore implements the Nodeinfostore interface. BoltStore uses the embedded
// bolt database to persist data to disc.
type BoltStore struct {
	db               *bolt.DB
	bucket           *bolt.Bucket
	onlineStatusJob  *scheduler.ScheduledJob
	gwOfflineHandler []func(string)
}

type JsonBool struct {
	Value bool
}

const NodeinfoBucket string = "nodeinfos"
const StatisticsBucket string = "statistics"
const StatusInfoBucket string = "statusinfo"
const NeighboursBucket string = "neighbours"
const GatewayBucket string = "gateways"

var AllBucketNames = []string{NodeinfoBucket, StatisticsBucket,
	StatusInfoBucket, NeighboursBucket, GatewayBucket}

// NewBoltStore creates a new BoltStore where the database file is located at
// the given path. If the file does not exist it will be created. If there is
// already a bolt database at the given path this BoltStore will contain its data.
func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &BoltStore{db: db}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range AllBucketNames {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	store.gwOfflineHandler = make([]func(string), 0, 10)
	store.onlineStatusJob = scheduler.NewJob(time.Minute*1, store.calculateOnlineStatus, false)
	return store, nil
}

// Close closes the underlying bolt database.
func (b *BoltStore) Close() {
	b.onlineStatusJob.Stop()
	b.db.Close()
}

func (bs *BoltStore) calculateOnlineStatus() {
	now := time.Now()
	updateInterval := conf.UInt("announced.interval.statistics", 300)
	offlineInterval := updateInterval * 2
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(StatusInfoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			status := NodeStatusInfo{}
			err := json.Unmarshal(v, &status)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"nodeId":     k,
					"jsonString": string(v),
				}).Error("Can't unmarshall json from node status info")
				continue
			}
			lastseen, _ := time.Parse(TimeFormat, status.Lastseen)
			if (now.Unix() - lastseen.Unix()) > int64(offlineInterval) {
				status.Online = false
				for _, handler := range bs.gwOfflineHandler {
					go handler(string(k))
				}
				data, err := json.Marshal(status)
				if err != nil {
					b.Put(k, data)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error in database transaction while updating online status")
	}
}

func (b *BoltStore) put(key, bucket string, data interface{}) {
	bytes, err := json.Marshal(data)
	if err == nil && bytes != nil {
		err = b.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			if b == nil {
				return fmt.Errorf("Bucket %s was null", bucket)
			}
			err = b.Put([]byte(key), bytes)
			return err
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
		if v == nil {
			return fmt.Errorf("Can't find object for key %s", key)
		}
		err := json.Unmarshal(v, object)
		return err
	})
	return err
}

func (b *BoltStore) NotifyNodeOffline(handler func(string)) {
	b.gwOfflineHandler = append(b.gwOfflineHandler, handler)
}

func (b *BoltStore) GetNodeInfo(nodeId string) (NodeInfo, error) {
	info := &NodeInfo{}
	err := b.get(nodeId, NodeinfoBucket, info)
	if err != nil {
		return NodeInfo{}, err
	}
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
			allNodeinfos = append(allNodeinfos, *nodeinfo)
		}
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": NodeinfoBucket,
		}).Error("Error iterating over all values")
	}
	return allNodeinfos
}

func (b *BoltStore) GetStatistics(nodeId string) (StatisticsStruct, error) {
	statistics := &StatisticsStruct{}
	err := b.get(nodeId, StatisticsBucket, statistics)
	if err != nil {
		return StatisticsStruct{}, err
	}
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
			allStatistics = append(allStatistics, *statistics)
		}
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": StatisticsBucket,
		}).Error("Error iterating over all values")
	}
	return allStatistics
}

func (b *BoltStore) GetNodeStatusInfo(nodeId string) (NodeStatusInfo, error) {
	info := &NodeStatusInfo{}
	err := b.get(nodeId, StatusInfoBucket, info)
	if err != nil {
		return NodeStatusInfo{}, err
	}
	if info.NodeId == "" {
		info.NodeId = nodeId
	}
	return *info, err
}

func (b *BoltStore) PutNodeStatusInfo(nodeId string, info NodeStatusInfo) {
	if info.NodeId == "" {
		info.NodeId = nodeId
	}
	b.put(nodeId, StatusInfoBucket, info)
}

func (b *BoltStore) GetNodeStatusInfos() []NodeStatusInfo {
	allStatusInfos := make([]NodeStatusInfo, 0, 500)
	err := b.allValues(StatusInfoBucket, func(key string, data []byte) {
		status := &NodeStatusInfo{}
		err := json.Unmarshal(data, status)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"jsonString": string(data),
				"nodeid":     key,
			}).Error("Error unmarshalling json data")
		} else {
			if status.NodeId == "" {
				status.NodeId = key
			}
			allStatusInfos = append(allStatusInfos, *status)
		}
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": StatusInfoBucket,
		}).Error("Error iterating over all values")
	}
	return allStatusInfos
}

func (b *BoltStore) GetNodeNeighbours(nodeId string) (NeighbourStruct, error) {
	neighbours := &NeighbourStruct{}
	err := b.get(nodeId, NeighboursBucket, neighbours)
	if err != nil {
		return NeighbourStruct{}, nil
	}
	return *neighbours, err
}

func (b *BoltStore) PutNodeNeighbours(neighbours NeighbourStruct) {
	b.put(neighbours.NodeId, NeighboursBucket, neighbours)
}

func (b *BoltStore) GetAllNeighbours() []NeighbourStruct {
	allNeighbours := make([]NeighbourStruct, 0, 500)
	err := b.allValues(NeighboursBucket, func(key string, data []byte) {
		neighbours := &NeighbourStruct{}
		err := json.Unmarshal(data, neighbours)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"jsonString": string(data),
				"nodeid":     key,
			}).Error("Error unmarshalling json data")
		} else {
			allNeighbours = append(allNeighbours, *neighbours)
		}
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"bucket": NeighboursBucket,
		}).Error("Error iterating over all values")
	}
	return allNeighbours
}

func (b *BoltStore) PutGateway(mac string) {
	b.put(mac, GatewayBucket, JsonBool{Value: true})
}

func (b *BoltStore) IsGateway(mac string) bool {
	result := &JsonBool{}
	err := b.get(mac, GatewayBucket, result)
	return err != nil && result.Value
}

func (b *BoltStore) RemoveGateway(mac string) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(GatewayBucket))
		err := b.Delete([]byte(mac))
		return err
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err,
			"gateway-mac": mac,
		}).Error("Error deleting gateway from bolt store")
	}
}
