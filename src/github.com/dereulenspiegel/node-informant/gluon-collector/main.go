package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

type LogPipe struct {
	logFile *bufio.Writer
}

type FakeGraphJson struct{}

func (f *FakeGraphJson) FakeGraphJson(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("/srv/ffmap-data/graph.json")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error opening graph.json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error reading graph.json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(dataBytes)
}

func (f *FakeGraphJson) Routes() []httpserver.Route {
	fakeRoutes := []httpserver.Route{
		httpserver.Route{"GraphJson", "GET", "/graph.json", f.FakeGraphJson},
	}
	return fakeRoutes
}

type FakeNodesJson struct{}

func (f *FakeNodesJson) FakeNodesJson(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("/srv/ffmap-data/nodes.json")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error opening nodes.json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error reading nodes.json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(dataBytes)
}

func (f *FakeNodesJson) Routes() []httpserver.Route {
	fakeRoutes := []httpserver.Route{
		httpserver.Route{"NodesJson", "GET", "/nodes.json", f.FakeNodesJson},
	}
	return fakeRoutes
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
	receivePipeline := pipeline.NewReceivePipeline(&pipeline.JsonParsePipe{}, &pipeline.DeflatePipe{})
	processPipe := pipeline.NewProcessPipeline(&pipeline.GatewayCollector{Store: store},
		&pipeline.NodeinfoCollector{Store: store}, &pipeline.StatisticsCollector{Store: store},
		&pipeline.NeighbourInfoCollector{Store: store}, &pipeline.StatusInfoCollector{Store: store})
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
		loader := &meshviewer.DataLoader{Store: store}
		err = loader.LoadNodesFromFile(nodesJsonPath)
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

	graphGenerator := &meshviewer.GraphGenerator{Store: store}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: store}

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

	httpserver.StartHttpServerBlocking(store, graphGenerator, nodesGenerator)
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
