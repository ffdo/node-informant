package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/httpserver"
)

var ifaceName = flag.String("iface", "", "Interface")
var udpPort = flag.Int("udpPort", 12444, "UDP Port")
var httpPort = flag.Int("httpPort", 8080, "Http Port")
var nodeinfoInterval = flag.Int("infoInterval", 600, "Interval between nodeinfo requests")
var statisticsInterval = flag.Int("statisticsInterval", 300, "Interval between statistics requests")
var nodesJsonPath = flag.String("nodesjson", "", "Static file with node information")
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

func Assemble(iface string, announcedPort int, httpListenAddr string) error {
	log.Printf("Assembling everything, creating requester")
	requester, err := announced.NewRequester(iface, announcedPort)
	if err != nil {
		return err
	}
	log.Printf("Creating store and pipes")
	store := data.NewSimpleInMemoryStore()
	if *nodesJsonPath != "" {
		store.LoadNodesFromFile(*nodesJsonPath)
	}
	receivePipeline := data.NewReceivePipeline(&data.JsonParsePipe{}, &data.DeflatePipe{})
	processPipe := data.NewProcessPipeline(store)
	log.Printf("Adding process pipe end")
	go func() {
		processPipe.Dequeue(func(response data.ParsedResponse) {
			//Do nothing. This is the last step and we do not need to do anything here
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
	log.Printf("Setting up request timer")
	nodeinfoTimer := time.NewTicker(time.Second * time.Duration(*nodeinfoInterval))
	statisticsTimer := time.NewTicker(time.Second * time.Duration(*statisticsInterval))
	updateNodesJsonTimer := time.NewTicker(time.Minute * 1)
	quitChan := make(chan bool)
	go func() {
		for {
			select {
			case <-nodeinfoTimer.C:
				requester.Query("GET nodeinfo")
			case <-statisticsTimer.C:
				requester.Query("GET statistics")
			case <-updateNodesJsonTimer.C:
				store.UpdateNodesJson()
			case <-quitChan:
				nodeinfoTimer.Stop()
			}
		}
	}()
	requester.Query("GET nodeinfo")
	requester.Query("GET statistics")
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
	conf.ParseConfig(*configFilePath)
	ConfigureLogger()
	err := Assemble(*ifaceName, *udpPort, fmt.Sprintf(":%d", *httpPort))
	log.Errorf("Error assembling application: %v", err)
}
