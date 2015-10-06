package main

import (
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/meshviewer"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

type TestDataReceiver struct {
	TestData []announced.Response
}

func (t *TestDataReceiver) Receive(rFunc func(announced.Response)) {
	for _, data := range t.TestData {
		rFunc(data)
	}
}

func executeCompletePipe(t *testing.T, store data.Nodeinfostore) {
	log.SetLevel(log.ErrorLevel)
	assert := assert.New(t)
	testReceiver := &TestDataReceiver{TestData: test.TestData}

	i := 0
	closeables, err := BuildPipelines(store, testReceiver, func(response data.ParsedResponse) {
		i = i + 1
	})
	assert.Nil(err)

	time.Sleep(time.Second * 2)

	for _, closable := range closeables {
		closable.Close()
	}

	assert.Equal(len(test.TestData), i)

	graphGenerator := &meshviewer.GraphGenerator{Store: store}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: store}
	graph := graphGenerator.GenerateGraph()
	assert.NotNil(graph)
	assert.Equal(232, len(graph.Batadv.Nodes))
	assert.Equal(72, len(graph.Batadv.Links))

	nodes := nodesGenerator.GetNodesJson()
	assert.NotNil(nodes)
}

func TestCompletePipe(t *testing.T) {
	store := data.NewSimpleInMemoryStore()
	executeCompletePipe(t, store)
}

func TestCompletePipeWithBoltStore(t *testing.T) {
	assert := assert.New(t)
	dbPath := "./bolt.db"
	defer os.RemoveAll(dbPath)
	store, err := data.NewBoltStore(dbPath)
	assert.Nil(err)
	assert.NotNil(store)
	executeCompletePipe(t, store)
	store.Close()
}
