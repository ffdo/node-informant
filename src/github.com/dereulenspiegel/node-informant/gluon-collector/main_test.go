package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/meshviewer"
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

type TestDataReceiver struct {
	TestData []announced.Response
}

func (t *TestDataReceiver) Receive(rFunc func(announced.Response)) {
	for _, data := range t.TestData {
		rFunc(data)
	}
}

func TestCompletePipe(t *testing.T) {
	log.SetLevel(log.WarnLevel)
	assert := assert.New(t)
	err := LoadTestData()
	assert.Nil(err)
	testReceiver := &TestDataReceiver{TestData: testData}
	store := data.NewSimpleInMemoryStore()

	i := 0
	closeables, err := BuildPipelines(store, testReceiver, func(response data.ParsedResponse) {
		i = i + 1
	})

	time.Sleep(time.Second * 2)

	for _, closable := range closeables {
		closable.Close()
	}

	assert.Equal(len(testData), i)

	graphGenerator := &meshviewer.GraphGenerator{Store: store}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: store}
	graph := graphGenerator.GenerateGraph()
	assert.NotNil(graph)
	assert.Equal(232, len(graph.Batadv.Nodes))
	assert.Equal(133, len(graph.Batadv.Links))

	nodes := nodesGenerator.GetNodesJson()
	assert.NotNil(nodes)
}
