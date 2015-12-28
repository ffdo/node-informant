package collector

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/data"
)

type CollectionFrequency int

const (
	CollectOnce CollectionFrequency = -1
	DontPoll    CollectionFrequency = -2
)

type FactCollector interface {
	Init(map[string]interface{}) error
	Collect() (interface{}, error)
	Path() string
	Frequency() CollectionFrequency
}

type CreateFactCollector func() FactCollector

var (
	factCollectorCreators []CreateFactCollector
	factCollectors        []FactCollector

	collectionJobs []*scheduler.ScheduledJob
)

func init() {
	factCollectorCreators = make([]CreateFactCollector, 0, 20)
	factCollectors = make([]FactCollector, 0, 20)
	collectionJobs = make([]*scheduler.ScheduledJob, 0, 20)
}

func Register(createColletorFunc CreateFactCollector) {
	factCollectorCreators = append(factCollectorCreators, createColletorFunc)
}

func collectFacts(collector FactCollector) {
	path := collector.Path()
	fact, err := collector.Collect()
	if err == nil {
		data.MergeCollectedData(path, fact)
	} else {
		log.WithFields(log.Fields{
			"collectorPath": path,
			"error":         err,
		}).Error("Can't collect metric")
	}
}

func isCollectorEnabed(path string, config map[string]interface{}) (bool, map[string]interface{}) {
	if subConfig, exists := config[path]; exists {
		collectorConfig := subConfig.(map[string]interface{})
		if enabledValue, enabledExists := collectorConfig["enabled"]; enabledExists {
			if enabledValue.(bool) {
				return true, collectorConfig
			}
		}
	}
	return false, nil
}

func InitCollection(collectorConfig map[string]interface{}) {
	for _, creator := range factCollectorCreators {
		factCollector := creator()
		factCollectors = append(factCollectors, factCollector)
		path := factCollector.Path()
		if isEnabled, config := isCollectorEnabed(path, collectorConfig); isEnabled {
			if err := factCollector.Init(config); err == nil {
				frequency := factCollector.Frequency()
				if frequency == CollectOnce {
					collectFacts(factCollector)
				} else if frequency > 0 {
					func(factCollector FactCollector) {
						collectJob := scheduler.NewJob(time.Second*time.Duration(frequency), func() {
							collectFacts(factCollector)
						}, true)
						collectionJobs = append(collectionJobs, collectJob)
					}(factCollector)
				}
			} else {
				log.WithFields(log.Fields{
					"collectorPath": path,
					"error":         err,
				}).Error("Can't initialise collector")
			}
		}
	}
}
