package main

import (
	"testing"

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

func TestPrometheusClientCounter(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)

	// set initial value
	initialCount := 10
	clientCounts := []int{13, 11, 15}
	prometheus.TotalClientCounter.Set(float64(initialCount))

	store := data.NewSimpleInMemoryStore()

	readers := []data.ParsedResponseReader{
		prometheus.ClientCounter,
		collectors.StatisticsCollector,
	}

	// push data into the readers
	for step, clientCount := range clientCounts {
		packet := data.StatisticsResponse{
			Statistics: &data.StatisticsStruct{
				NodeId: testNodeId,
				Clients: data.ClientStatistics{
					Wifi:  clientCount,
					Total: clientCount,
				},
			},
		}

		pipeline.FeedParsedResponseReaders(readers, store, packet)

		// check value
		value := collectGaugeValue(prometheus.TotalClientCounter)
		assert.Equal(clientCounts[step]+initialCount, int(value))
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
