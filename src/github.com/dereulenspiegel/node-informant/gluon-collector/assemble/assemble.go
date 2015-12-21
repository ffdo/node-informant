package assemble

import (
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/collectors"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/pipeline"
	"github.com/dereulenspiegel/node-informant/gluon-collector/prometheus"
)

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
