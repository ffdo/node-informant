package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/api"
	"github.com/dereulenspiegel/node-informant/gluon-collector/collectors"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
	"github.com/dereulenspiegel/node-informant/gluon-collector/meshviewer"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	"github.com/dereulenspiegel/node-informant/gluon-collector/prometheus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/scheduler"
)

var importPath = flag.String("import", "", "Import data from this path")
var importType = flag.String("importType", "ffmap-backend", "The data format to import from, i.e ffmap-backend")

var DataStore data.Nodeinfostore
var Closeables []io.Closer

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

func getProcessPipes(store data.Nodeinfostore) []pipeline.ProcessPipe {
	pipes := make([]pipeline.ProcessPipe, 0, 10)

	pipes = append(pipes, prometheus.GetPrometheusProcessPipes(store)...)
	pipes = append(pipes, &collectors.GatewayCollector{Store: store},
		&collectors.NodeinfoCollector{Store: store}, &collectors.StatisticsCollector{Store: store},
		&collectors.NeighbourInfoCollector{Store: store}, &collectors.StatusInfoCollector{Store: store})
	return pipes
}

func BuildPipelines(store data.Nodeinfostore, receiver announced.AnnouncedPacketReceiver, pipeEnd func(response data.ParsedResponse)) ([]io.Closer, error) {

	closeables := make([]io.Closer, 0, 2)

	receivePipeline := pipeline.NewReceivePipeline(&pipeline.JsonParsePipe{}, &pipeline.DeflatePipe{})
	processPipe := pipeline.NewProcessPipeline(getProcessPipes(store)...)
	closeables = append(closeables, receivePipeline, processPipe)
	log.Printf("Adding process pipe end")
	go func() {
		processPipe.Dequeue(pipeEnd)
	}()
	log.Printf("Connecting requester to receive pipeline")
	go func() {
		receiver.Receive(func(response announced.Response) {
			receivePipeline.Enqueue(response)
		})
	}()
	log.Printf("Connecting receive to process pipeline")
	//Connect the receive to the process pipeline
	go func() {
		receivePipeline.Dequeue(func(response data.ParsedResponse) {
			processPipe.Enqueue(response)
		})
	}()
	return closeables, nil
}

func Assemble() ([]io.Closer, error) {
	requester := buildReceiver()
	closeables, err := BuildPipelines(DataStore, requester, func(response data.ParsedResponse) {
		//Do nothing. This is the last step and we do not need to do anything here,
		// just pull the chan clean
	})
	closeables = append(closeables, requester)
	if err != nil {
		return closeables, err
	}
	graphGenerator := &meshviewer.GraphGenerator{Store: DataStore}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: DataStore}
	missingUpdate := &MissingUpdater{Store: DataStore, Requester: requester}
	DataStore.NotifyNodeOffline(missingUpdate.CheckNodeUnicast)
	nodesGenerator.UpdateNodesJson()
	graphGenerator.UpdateGraphJson()

	log.Printf("Setting up request timer")
	nodeinfoInterval := conf.Global.UInt("announced.interval.nodeinfo", 1800)
	statisticsInterval := conf.Global.UInt("announced.interval.statistics", 300)

	scheduler.NewJob(time.Second*time.Duration(nodeinfoInterval), func() {
		log.Debug("Querying Nodeinfos via Multicast")
		requester.Query("GET nodeinfo")
	}, true)
	time.Sleep(time.Second * 10)
	scheduler.NewJob(time.Second*time.Duration(statisticsInterval), func() {
		log.Debug("Querying statistics via Multicast")
		requester.Query("GET statistics")
		time.Sleep(time.Second * 25)
		log.Debug("Querying neighbours via Multicast")
		requester.Query("GET neighbours")
	}, true)

	scheduler.NewJob(time.Minute*1, func() {
		graphGenerator.UpdateGraphJson()
	}, false)

	scheduler.NewJob(time.Minute*1, func() {
		nodesGenerator.UpdateNodesJson()
	}, false)
	httpApi := &api.HttpApi{Store: DataStore}
	httpserver.StartHttpServerBlocking(httpApi, graphGenerator, nodesGenerator)
	return closeables, nil
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
		if err != nil && os.IsNotExist(err) {
			file, err = os.Create(filePath)
		}
		if err != nil {
			log.WithFields(log.Fields{
				"err":         err,
				"logFilePath": filePath,
			}).Fatal("Can't open/create logfile")
		} else {
			log.SetOutput(file)
		}
	}
}

func CreateDataStore() {
	dbType := conf.UString("store.type", "memory")
	switch dbType {
	case "memory":
		DataStore = data.NewSimpleInMemoryStore()
	case "bolt":
		storagePath := conf.UString("store.path", "/opt/gluon-collector/collector.db")
		boltStore, err := data.NewBoltStore(storagePath)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"storePath": storagePath,
				"storeType": dbType,
			}).Fatal("Can't create bolt store")
		} else {
			Closeables = append(Closeables, boltStore)
			DataStore = boltStore
		}
	default:
		log.Fatalf("Unknown store type %s", dbType)
	}
}

func Stop() {
	for _, c := range Closeables {
		c.Close()
	}
}

func ListenToSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		for _ = range c {
			log.Print("Shutting down...")
			Stop()
		}
	}()
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
	Closeables = make([]io.Closer, 0, 5)
	flag.Parse()
	conf.InitConfig()
	if conf.Global == nil {
		log.Fatal("Configuration couldn't be parsed")
	}
	ConfigureLogger()
	CreateDataStore()
	prometheus.ProcessStoredValues(DataStore)
	if *importPath != "" {
		ImportData()
	}
	closeables, err := Assemble()
	Closeables = append(Closeables, closeables...)
	ListenToSig()
	log.Errorf("Error assembling application: %v", err)
}
