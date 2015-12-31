package assemble

import (
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/ffdo/node-informant/announced"
	"github.com/ffdo/node-informant/gluon-collector/collectors"
	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/pipeline"
	"github.com/ffdo/node-informant/gluon-collector/prometheus"
)

func BuildPipelines(store data.Nodeinfostore, receiver announced.AnnouncedPacketReceiver, pipeEnd func(response data.ParsedResponse)) ([]io.Closer, error) {

	receivePipeline := pipeline.NewReceivePipeline(&pipeline.JsonParsePipe{}, &pipeline.DeflatePipe{})
	responseReaders := append(prometheus.GetPrometheusProcessPipes(),
		collectors.GatewayCollector,
		collectors.NodeinfoCollector,
		collectors.StatisticsCollector,
		collectors.NeighbourInfoCollector,
		collectors.StatusInfoCollector,
	)

	log.Printf("Connecting requester to receive pipeline")
	go func() {
		receiver.Receive(func(response announced.Response) {
			receivePipeline.Enqueue(response)
		})
	}()

	log.Printf("Connecting receive to process pipeline")
	go func() {
		receivePipeline.Dequeue(func(response data.ParsedResponse) {
			pipeline.FeedParsedResponseReaders(responseReaders, store, response)
			pipeEnd(response)
		})
	}()

	return []io.Closer{receivePipeline}, nil
}
