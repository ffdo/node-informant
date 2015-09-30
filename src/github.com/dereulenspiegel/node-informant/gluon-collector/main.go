package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")

type LogPipe struct {
	logFile *bufio.Writer
}

func NewLogPipe() *LogPipe {
	logFile, err := os.Create("/opt/rawdata.log")
	if err != nil {
		log.Fatalf("Cannot open raw data logfile")
	}
	writer := bufio.NewWriter(logFile)
	return &LogPipe{logFile: writer}
}

func (l *LogPipe) Process(in chan announced.Response) chan announced.Response {
	out := make(chan announced.Response)
	go func() {
		for response := range in {
			_, err := l.logFile.WriteString(fmt.Sprintf("%s|", response.String()))
			l.logFile.Flush()
			if err != nil {
				log.Fatalf("Can't write to logfile: %v", err)
			}
			out <- response
		}
	}()
	return out
}

func BuildPipelines(requester announced.Requester, store *data.SimpleInMemoryStore) error {
	receivePipeline := data.NewReceivePipeline(&pipeline.JsonParsePipe{}, &pipeline.DeflatePipe{})
	processPipe := data.NewProcessPipeline(&pipeline.GatewayCollector{Store: store},
		&pipeline.NodeinfoCollector{Store: store}, &pipeline.StatisticsCollector{Store: store},
		&pipeline.NeighbourInfoCollector{Store: store})
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

	scheduler.NewJob(time.Second*time.Duration(nodeinfoInterval), func() {
		requester.Query("GET nodeinfo")
	}, true)

	scheduler.NewJob(time.Second*time.Duration(statisticsInterval), func() {
		requester.Query("GET statistics")
	}, true)

	scheduler.NewJob(time.Second*time.Duration(statisticsInterval), func() {
		requester.Query("GET neighbours")
	}, true)

	scheduler.NewJob(time.Minute*1, func() {
		store.UpdateNodesJson()
	}, true)

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
