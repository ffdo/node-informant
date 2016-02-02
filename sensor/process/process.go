package process

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/sensor/collector"
	"github.com/ffdo/node-informant/sensor/data"
	"github.com/ffdo/node-informant/sensor/module"
	"github.com/olebedev/config"
)

type ProcessFunction func(data.NodeData) error

var (
	processFunctions []ProcessFunction

	ProcessWaitGroup *sync.WaitGroup
	processCount     int
)

func init() {
	processFunctions = make([]ProcessFunction, 0, 10)
	ProcessWaitGroup = &sync.WaitGroup{}
	module.Register(module.Module{
		Init:  Init,
		Start: Start,
		Close: Close,
		Name:  "process",
	})
}

func Init(cfg *config.Config) error {
	processCount = cfg.UInt("process.count", 2)
	return nil
}

func Start() error {
	log.Debug("Really starting process module")
	receiveChan := make(chan data.NodeData, 100)
	log.Debug("Starting receive on collector")
	collector.Receive(receiveChan)
	for i := 0; i < processCount; i++ {
		log.Debugf("Startinng processor %d", i)
		ProcessWaitGroup.Add(1)
		go Receive(receiveChan)
	}
	return nil
}

func Close() error {
	return nil
}

func RegisterProcessFunction(function ProcessFunction) {
	processFunctions = append(processFunctions, function)
}

func Receive(in chan data.NodeData) {
	defer ProcessWaitGroup.Done()
	for packet := range in {
		go func(packet data.NodeData) {
			for _, function := range processFunctions {
				if err := function(packet); err != nil {
					// TODO Log this error
					break
				}
			}
		}(packet)
	}
}
