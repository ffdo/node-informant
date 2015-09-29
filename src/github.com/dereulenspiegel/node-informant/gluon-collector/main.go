package main

import (
	"flag"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")

type LogPipe struct {
}

func (l *LogPipe) Process(in chan announced.Response) chan announced.Response {
	out := make(chan announced.Response)
	go func() {
		for response := range in {
			log.Debugf("Received packet from %#v: %#v", response.ClientAddr, response.Payload)
			out <- response
		}
	}()
	return out
}

func BuildPipelines(requester announced.Requester, store *data.SimpleInMemoryStore) error {
	receivePipeline := data.NewReceivePipeline(&data.JsonParsePipe{}, &data.DeflatePipe{})
	processPipe := data.NewProcessPipeline(&data.GatewayCollector{Store: store},
		&data.NodeinfoCollector{Store: store}, &data.StatisticsCollector{Store: store},
		&data.NeighbourInfoCollector{Store: store})
	log.Printf("Adding process pipe end")
	go func() {
		processPipe.Dequeue(func(response data.ParsedResponse) {
			//Do nothing. This is the last step and we do not need to do anything here,
			// just pull the chan clean
		})
	}()
	log.Printf("Connecting requester to receive pipeline")
	go func() {
		for response := range requester.ReceiveChan {
			receivePipeline.Enqueue(response)
		}
	}()
	log.Printf("Connecting receive to process pipeline")
	//Connect the receive to the process pipeline
	go func() {
		receivePipeline.Dequeue(func(response data.ParsedResponse) {
			processPipe.Enqueue(response)
		})
	}()
	return nil
}

func Assemble() error {
	store := data.NewSimpleInMemoryStore()
	nodesJsonPath, err := conf.Global.String("nodesJsonPath")
	if err != nil {
		log.Infof("Loading node information from file %s", nodesJsonPath)
		err = store.LoadNodesFromFile(nodesJsonPath)
		log.WithFields(log.Fields{
			"error": err,
			"path":  nodesJsonPath,
		}).Error("Can't node information from file")
	}
	iface, err := conf.Global.String("announced.interface")
	if err != nil {
		return err
	}
	requester, err := announced.NewRequester(iface, conf.Global.UInt("announced.port", 12444))
	if err != nil {
		log.Fatalf("Can't create requester: %v", err)
		return err
	}
	err = BuildPipelines(requester, store)

	log.Printf("Setting up request timer")
	nodeinfoInterval := conf.Global.UInt("announced.interval.nodeinfo", 1800)
	statisticsInterval := conf.Global.UInt("announced.interval.statistics", 300)

	nodeinfoTimer := time.NewTicker(time.Second * time.Duration(nodeinfoInterval))
	statisticsTimer := time.NewTicker(time.Second * time.Duration(statisticsInterval))
	neighbourTimer := time.NewTicker(time.Second * time.Duration(statisticsInterval))
	updateNodesJsonTimer := time.NewTicker(time.Minute * 1)
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-nodeinfoTimer.C:
				requester.Query("GET nodeinfo")
			case <-statisticsTimer.C:
				requester.Query("GET statistics")
			case <-neighbourTimer.C:
				requester.Query("GET neighbours")
			case <-updateNodesJsonTimer.C:
				store.UpdateNodesJson()
			case <-quitChan:
				nodeinfoTimer.Stop()
				updateNodesJsonTimer.Stop()
				statisticsTimer.Stop()
				neighbourTimer.Stop()
			}
		}
	}()
	requester.Query("GET nodeinfo")
	requester.Query("GET statistics")
	requester.Query("GET neighbours")
	httpserver.StartHttpServerBlocking(store)
	return nil
}

func ConfigureLogger() {
	lvl, err := log.ParseLevel(conf.Global.UString("logger.level", "debug"))
	if err != nil {
		lvl = log.DebugLevel
		log.Errorf("Error while parsing log level, falling back to Debug: %v", err)
	}
	log.SetLevel(lvl)
}

func main() {
	flag.Parse()
	err := conf.ParseConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Error parsing config file %s: %v", *configFilePath, err)
	}
	ConfigureLogger()
	err = Assemble()
	log.Errorf("Error assembling application: %v", err)
}
