package main

import (
	"testing"
	"time"

	"github.com/ffdo/node-informant/gluon-collector/collectors"
	"github.com/ffdo/node-informant/gluon-collector/config"
	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/pipeline"
	"github.com/ffdo/node-informant/gluon-collector/prometheus"
	"github.com/ffdo/node-informant/gluon-collector/test"
	stat "github.com/prometheus/client_golang/prometheus"

	dto "github.com/prometheus/client_model/go"

	"github.com/stretchr/testify/assert"
)

var testNodeId string = "e8de27252554"

func collectGaugeValue(gauge stat.Collector) float64 {
	collectChan := make(chan stat.Metric)
	var value float64
	go gauge.Collect(collectChan)
	select {
	case metric := <-collectChan:
		metricDto := &dto.Metric{}
		metric.Write(metricDto)
		value = metricDto.GetGauge().GetValue()
	}
	close(collectChan)
	return value
}

func feedClientsStat(processPipeline *pipeline.ProcessPipeline, clientCount int) {
	clients1 := data.ClientStatistics{
		Wifi:  clientCount,
		Total: clientCount,
	}
	stats1 := data.StatisticsStruct{
		Clients: clients1,
		NodeId:  testNodeId,
	}

	packet1 := data.StatisticsResponse{
		Statistics: &stats1,
	}

	processPipeline.Enqueue(packet1)
}

func TestPrometheusClientCounter(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)

	var expectedClientCounts = []float64{13, 11, 15}
	finishChan := make(chan bool)

	store := data.NewSimpleInMemoryStore()
	processPipeline := pipeline.NewProcessPipeline(&prometheus.ClientCountPipe{Store: store},
		&collectors.StatisticsCollector{Store: store})

	prometheus.TotalClientCounter.Set(10.0)

	var packetCount int = 0
	go processPipeline.Dequeue(func(response data.ParsedResponse) {
		value := collectGaugeValue(prometheus.TotalClientCounter)
		assert.Equal(expectedClientCounts[packetCount], value)
		packetCount = packetCount + 1
		if packetCount == len(expectedClientCounts) {
			finishChan <- true
			close(finishChan)
		}
	})

	feedClientsStat(processPipeline, 3)
	// Give the collector pipe a little time to execute its go routin
	// in production it is very very unrealistic that we will have two Statistics
	// Responses from the same node in the channel at the same time.
	time.Sleep(time.Millisecond * 50)
	feedClientsStat(processPipeline, 1)
	time.Sleep(time.Millisecond * 50)
	feedClientsStat(processPipeline, 5)

	for range finishChan {
		processPipeline.Close()
	}
}

func TestPrometheusClientCounterWithFullLabels(t *testing.T) {
	config.Global.Set("prometheus.namelabel", true)
	config.Global.Set("prometheus.sitecodelabel", true)
	store := data.NewSimpleInMemoryStore()
	test.ExecuteCompletePipe(t, store)

	// It is kinda complicated to collect values from CounterVecs,
	// so for now this test simply makes sure that the application doesn't
	// crash if extended node labels are activated.
}
