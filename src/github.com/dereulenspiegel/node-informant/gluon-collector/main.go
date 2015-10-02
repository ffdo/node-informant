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
	"github.com/dereulenspiegel/node-informant/gluon-collector/meshviewer"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")
var importPath = flag.String("import", "", "Import data from this path")
var importType = flag.String("importType", "", "The data format to import from, i.e ffmap-backend")

var DataStore data.Nodeinfostore

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

func BuildPipelines(requester announced.Requester) error {
	receivePipeline := pipeline.NewReceivePipeline(&pipeline.JsonParsePipe{}, &pipeline.DeflatePipe{})
	processPipe := pipeline.NewProcessPipeline(&pipeline.GatewayCollector{Store: DataStore},
		&pipeline.NodeinfoCollector{Store: DataStore}, &pipeline.StatisticsCollector{Store: DataStore},
		&pipeline.NeighbourInfoCollector{Store: DataStore}, &pipeline.StatusInfoCollector{Store: DataStore})
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
	iface, err := conf.Global.String("announced.interface")
	if err != nil {
		return err
	}
	requester, err := announced.NewRequester(iface, conf.Global.UInt("announced.port", 12444))
	if err != nil {
		log.Fatalf("Can't create requester: %v", err)
		return err
	}
	err = BuildPipelines(requester)

	graphGenerator := &meshviewer.GraphGenerator{Store: DataStore}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: DataStore}

	log.Printf("Setting up request timer")
	nodeinfoInterval := conf.Global.UInt("announced.interval.nodeinfo", 1800)
	statisticsInterval := conf.Global.UInt("announced.interval.statistics", 300)

	scheduler.NewJob(time.Second*time.Duration(nodeinfoInterval), func() {
		requester.Query("GET nodeinfo")
	}, true)
	time.Sleep(time.Second * 10)
	scheduler.NewJob(time.Second*time.Duration(statisticsInterval), func() {
		requester.Query("GET statistics")
		time.Sleep(time.Second * 10)
		requester.Query("GET neighbours")
	}, true)

	scheduler.NewJob(time.Minute*1, func() {
		graphGenerator.UpdateGraphJson()
	}, false)

	scheduler.NewJob(time.Minute*1, func() {
		nodesGenerator.UpdateNodesJson()
	}, false)

	httpserver.StartHttpServerBlocking(graphGenerator, nodesGenerator)
	return nil
}

func ConfigureLogger() {
	lvl, err := log.ParseLevel(conf.Global.UString("logger.level", "debug"))
	if err != nil {
		lvl = log.DebugLevel
		log.Errorf("Error while parsing log level, falling back to Debug: %v", err)
	}
	log.SetLevel(lvl)

	filePath, err := conf.Global.String("logger.file")
	if err == nil && filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"err":         err,
				"logFilePath": filePath,
			}).Fatal("Can't open logfile")
		} else {
			log.SetOutput(file)
		}
	}
}

func CreateDataStore() {
	DataStore = data.NewSimpleInMemoryStore()
}

func ImportData() {
	log.Infof("Loading node information from file %s", *importPath)
	// TODO choose DataLoader depending on importType
	loader := &meshviewer.FFMapBackendDataLoader{Store: DataStore}
	err := loader.LoadNodesFromFile(*importPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  *importPath,
		}).Error("Can't node information from file")
	}
}

func main() {
	flag.Parse()
	err := conf.ParseConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Error parsing config file %s: %v", *configFilePath, err)
	}
	ConfigureLogger()
	CreateDataStore()
	if *importPath != "" {
		ImportData()
	}
	err = Assemble()
	log.Errorf("Error assembling application: %v", err)
}
