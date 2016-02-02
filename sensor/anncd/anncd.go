package anncd

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/announced"
	"github.com/ffdo/node-informant/gluon-collector/scheduler"
	"github.com/ffdo/node-informant/sensor/collector"
	"github.com/ffdo/node-informant/sensor/data"
	"github.com/ffdo/node-informant/utils"
	"github.com/olebedev/config"
)

type processFunction func(*announced.Response) (*announced.Response, error)

type Config struct {
	RequestNodeInfoInterval   time.Duration
	RequestStatisticsInterval time.Duration
	RequestNeighbourInterval  time.Duration
	InterfaceName             string
	Port                      int
}

type anncdCollector struct {
	requester            *announced.Requester
	receiveChan          chan data.NodeData
	processFunctions     []processFunction
	config               Config
	requestNodeInfoJob   *scheduler.ScheduledJob
	requestStatisticsJob *scheduler.ScheduledJob
	requestNeighboursJob *scheduler.ScheduledJob
}

func init() {
	log.Info("Init of announced collector")
	collector.RegisterCollectorCreator("announced", NewCollector)
}

func createAnncdConfig(cfg *config.Config) (Config, error) {
	var err error
	anncdCfg := Config{}
	if anncdCfg.InterfaceName, err = cfg.String("interface"); err != nil {
		return anncdCfg, err
	}
	anncdCfg.Port = cfg.UInt("port", 12346)
	if anncdCfg.RequestNeighbourInterval, err = time.ParseDuration(cfg.UString("neighbourInterval", "5m")); err != nil {
		return anncdCfg, err
	}
	if anncdCfg.RequestNodeInfoInterval, err = time.ParseDuration(cfg.UString("nodeinfoInterval", "5m")); err != nil {
		return anncdCfg, err
	}
	if anncdCfg.RequestStatisticsInterval, err = time.ParseDuration(cfg.UString("statisticsInterval", "2m")); err != nil {
		return anncdCfg, err
	}
	return anncdCfg, nil
}

func NewCollector(cfg *config.Config) (collector.Collector, error) {
	anncdCfg, err := createAnncdConfig(cfg)
	if err != nil {
		return nil, err
	}
	coll := &anncdCollector{}
	coll.config = anncdCfg
	coll.processFunctions = make([]processFunction, 0, 1)
	coll.processFunctions = append(coll.processFunctions, deflatePayload)
	return coll, nil
}

func (a *anncdCollector) Close() error {
	a.requestNeighboursJob.Stop()
	a.requestNodeInfoJob.Stop()
	a.requestStatisticsJob.Stop()
	return a.requester.Close()
}

func (a *anncdCollector) Start() error {
	log.Debug("Starting announced collector")
	requester, err := announced.NewRequester(a.config.InterfaceName, a.config.Port)
	if err != nil {
		return err
	}
	a.requester = requester
	a.requestNeighboursJob = scheduler.NewJob(a.config.RequestNeighbourInterval,
		a.queryNeighbours, true)
	a.requestNodeInfoJob = scheduler.NewJob(a.config.RequestNodeInfoInterval,
		a.queryNodeinfo, true)
	a.requestStatisticsJob = scheduler.NewJob(a.config.RequestStatisticsInterval, a.queryStatistics, true)
	return nil
}

func (a *anncdCollector) queryNeighbours() {
	a.requester.Query("GET neighbours")
}

func (a *anncdCollector) queryNodeinfo() {
	a.requester.Query("GET nodeinfo")
}

func (a *anncdCollector) queryStatistics() {
	a.requester.Query("GET statistics")
}

func (a *anncdCollector) Receive(in chan data.NodeData) {
	a.receiveChan = in
	go a.requester.Receive(a.receiveRawPacket)
}

func (a *anncdCollector) receiveRawPacket(packet announced.Response) {
	go func(packet announced.Response) {
		currPacket := &packet
		var err error
		for _, function := range a.processFunctions {
			currPacket, err = function(currPacket)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"remoteAddress": packet.ClientAddr.String(),
					"payload":       packet.Payload,
				}).Error("Error while processing packet")
				break
			}
		}
		if err == nil {
			if nodeData, err := data.ParseJson(currPacket.Payload); err == nil {
				a.receiveChan <- nodeData
			} else {
				log.WithError(err).WithFields(log.Fields{
					"payloadString": string(currPacket.Payload),
					"remoteAddress": currPacket.ClientAddr.String(),
				}).Error("Error parsing json")
			}
		}
	}(packet)
}

func deflatePayload(packet *announced.Response) (*announced.Response, error) {
	decompressedData, err := utils.Deflate(packet.Payload)
	if err != nil {
		return nil, err
	}
	packet.Payload = decompressedData
	return packet, nil
}
