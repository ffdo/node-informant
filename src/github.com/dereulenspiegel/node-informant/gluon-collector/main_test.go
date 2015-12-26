package main

import (
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/meshviewer"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

func TestCompletePipe(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.DebugLevel)
	store := data.NewSimpleInMemoryStore()
	test.ExecuteCompletePipe(t, store)

	graphGenerator := &meshviewer.GraphGenerator{Store: store}
	nodesGenerator := &meshviewer.NodesJsonGenerator{Store: store}
	graph := graphGenerator.GenerateGraph()
	assert.NotNil(graph)
	assert.Equal(169, len(graph.Batadv.Nodes))
	assert.Equal(71, len(graph.Batadv.Links))

	nodes := nodesGenerator.GetNodesJson()
	assert.NotNil(nodes)
}

func TestCompletePipeWithBoltStore(t *testing.T) {
	assert := assert.New(t)
	log.SetLevel(log.DebugLevel)
	dbPath := "./bolt.db"
	defer os.RemoveAll(dbPath)
	store, err := data.NewBoltStore(dbPath)
	assert.Nil(err)
	assert.NotNil(store)
	test.ExecuteCompletePipe(t, store)
	store.Close()
}
