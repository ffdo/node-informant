package collector

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

type FactCollector interface {
	Init(map[string]interface{}) error
	Collect() interface{}
	Path() string
}

type CreateFactCollector func() FactCollector

var (
	FactCollectorCreators []CreateFactCollector
	FactCollectors        []FactCollector

	collectJob *scheduler.ScheduledJob
)

func init() {
	FactCollectorCreators = make([]CreateFactCollector, 0, 20)
	FactCollectors = make([]FactCollector, 0, 20)
}

func Register(createColletorFunc CreateFactCollector) {
	FactCollectorCreators = append(FactCollectorCreators, createColletorFunc)
}

func InitCollection(collectorConfig map[string]interface{}) {
	for _, creator := range FactCollectorCreators {
		FactCollector := creator(collectorConfig)
		if err := FactCollector.Init(); err != nil {
			FactCollectors = append(FactCollectors, FactCollector)
		} else {
			log.WithFields(log.Fields{
				"error":  err,
				"metric": FactCollector.Path(),
			}).Fatalf("Failed to initialize statistic collector")
		}
	}
	collectJob = scheduler.NewJob(time.Minute*1, func() {
		for _, collector := range FactCollectors {
			collector.Collect()
		}
	}, true)
}
