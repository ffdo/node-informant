package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/stretchr/testify/assert"
)

var testData []announced.Response

func stringToBytes(in string) []byte {
	parts := strings.Split(in, " ")
	bytes := make([]byte, len(parts))
	if len(parts) == 0 {
		log.Fatalf("No bytes to decode")
	}
	for i, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		y, err := strconv.Atoi(p)
		if err != nil {
			log.Fatalf("Can't convert string %s to int", p)
		}
		bytes[i] = byte(y)
	}
	return bytes
}

func LoadTestData() error {
	dataFile, err := os.Open("../../../../../testdata.raw")
	if err != nil {
		return err
	}
	dataBytes, err := ioutil.ReadAll(dataFile)
	if err != nil {
		return err
	}
	dataString := string(dataBytes)
	serializedPacktes := strings.Split(dataString, "|")
	responses := make([]announced.Response, 0, len(serializedPacktes))
	for _, text := range serializedPacktes {
		if strings.TrimSpace(text) == "" {
			continue
		}
		printableResponse := &announced.PrintableResponse{}
		err := json.Unmarshal([]byte(text), printableResponse)
		if err != nil {
			log.Printf("Error unmarshalling json %s: %v", text, err)
			return err
		}
		addr := &net.UDPAddr{
			IP:   net.ParseIP(printableResponse.Addr.IP),
			Port: printableResponse.Addr.Port,
			Zone: printableResponse.Addr.Zone,
		}
		response := announced.Response{
			ClientAddr: addr,
			Payload:    stringToBytes(printableResponse.Bytes),
		}
		responses = append(responses, response)
	}
	testData = responses
	log.Printf("Loaded %d packets", len(testData))
	return nil
}

func TestCompletePipe(t *testing.T) {
	assert := assert.New(t)
	err := LoadTestData()
	assert.Nil(err)

	store := data.NewSimpleInMemoryStore()
	receivePipeline := data.NewReceivePipeline(&data.JsonParsePipe{}, &data.DeflatePipe{})
	processPipe := data.NewProcessPipeline(&data.GatewayCollector{Store: store},
		&data.NodeinfoCollector{Store: store}, &data.StatisticsCollector{Store: store},
		&data.NeighbourInfoCollector{Store: store})

	i := 0
	go func() {
		processPipe.Dequeue(func(response data.ParsedResponse) {
			i = i + 1
		})
	}()

	//Connect the receive to the process pipeline
	go func() {
		receivePipeline.Dequeue(func(response data.ParsedResponse) {
			processPipe.Enqueue(response)
		})
	}()

	go func() {
		for _, response := range testData {
			receivePipeline.Enqueue(response)
		}
	}()

	time.Sleep(time.Second * 2)

	receivePipeline.Close()
	processPipe.Close()
	assert.Equal(len(testData), i)

	graphGenerator := &data.GraphGenerator{Store: store}
	graph := graphGenerator.GenerateGraphJson()
	assert.NotNil(graph)
	assert.Equal(232, len(graph.Batadv.Nodes))
	assert.Equal(11, len(graph.Batadv.Links))
}
