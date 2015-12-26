package collector

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

type StatCollector interface {
	Init(map[string]interface{}) error
	Collect() interface{}
	Path() string
}

type CreateStatCollector func() StatCollector

var (
	statCollectorCreators []CreateStatCollector
	statCollectors        []StatCollector

	collectJob *scheduler.ScheduledJob
)

func init() {
	statCollectorCreators = make([]CreateStatCollector, 0, 20)
	statCollectors = make([]StatCollector, 0, 20)
}

func Register(createColletorFunc CreateStatCollector) {
	statCollectorCreators = append(statCollectorCreators, createColletorFunc)
}

func InitCollection(collectorConfig map[string]interface{}) {
	for _, creator := range statCollectorCreators {
		statCollector := creator(collectorConfig)
		if err := statCollector.Init(); err != nil {
			statCollectors = append(statCollectors, statCollector)
		} else {
			log.WithFields(log.Fields{
				"error":  err,
				"metric": statCollector.Path(),
			}).Fatalf("Failed to initialize statistic collector")
		}
	}
	collectJob = scheduler.NewJob(time.Minute*1, func() {
		for _, collector := range statCollectors {
			collector.Collect()
		}
	}, true)
}
