package collector

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/sensor/data"
	"github.com/ffdo/node-informant/sensor/module"
	"github.com/olebedev/config"
)

var (
	CollectorCreators map[string]CollectorCreator
	parentCollector   Collector
)

func Init(cfg *config.Config) error {
	parentCollector = configureCollectors(cfg)
	return nil
}

func Start() error {
	return parentCollector.Start()
}

func Close() error {
	return parentCollector.Close()
}

func init() {
	CollectorCreators = make(map[string]CollectorCreator)
	module.Register(module.Module{
		Init:  Init,
		Start: Start,
		Close: Close,
		Name:  "collector",
	})
}

func configureCollectors(cfg *config.Config) Collector {
	collectorConfigs, err := cfg.List("collectors")
	if err != nil {
		log.WithError(err).Fatal("No collectors configured")
	}
	collectors := make([]Collector, 0, 5)
	for i, _ := range collectorConfigs {
		if collectorConfig, err := cfg.Get(fmt.Sprintf("collectors.%d", i)); err != nil {
			log.WithError(err).Fatal("Error when retrieving collector config")
		} else {
			if name, err := collectorConfig.String("type"); err != nil {
				log.WithError(err).Fatal("Type for collector is not specified")
			} else {
				log.WithField("collectorType", name).Info("Creating collector")
				creator := GetCollectorCreator(name)
				if coll, err := creator(collectorConfig); err != nil {
					log.WithError(err).WithField("collectorType", name).Fatal("Can't create collector")
				} else {
					collectors = append(collectors, coll)
				}
			}
		}
	}
	return MultiCollector(collectors...)
}

func RegisterCollectorCreator(name string, creator CollectorCreator) {
	log.WithField("collectorName", name).Debug("Registering collector")
	CollectorCreators[name] = creator
}

func GetCollectorCreator(name string) CollectorCreator {
	return CollectorCreators[name]
}

type Collector interface {
	io.Closer
	Start() error
	Receive(in chan data.NodeData)
}

type CollectorCreator func(*config.Config) (Collector, error)

type multiCollector struct {
	childs []Collector
}

func (m *multiCollector) Start() error {
	for _, child := range m.childs {
		if err := child.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiCollector) Close() error {
	for _, child := range m.childs {
		// TODO collect errors
		if err := child.Close(); err != nil {
			log.WithError(err).Error("Error closing collector")
		}
	}
	return nil
}

func (m *multiCollector) Receive(in chan data.NodeData) {
	for _, child := range m.childs {
		go child.Receive(in)
	}
}

func MultiCollector(collectors ...Collector) Collector {
	childs := make([]Collector, 0, len(collectors))
	childs = append(childs, collectors...)
	return &multiCollector{
		childs: childs,
	}
}

func Receive(in chan data.NodeData) {
	parentCollector.Receive(in)
}
