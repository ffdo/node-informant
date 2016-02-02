package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleMerge(t *testing.T) {
	assert := assert.New(t)

	nodeData, err := ParseJson([]byte(exampleNodeInfo))
	assert.Nil(err)
	newNodeData := NodeData{Root: make(map[string]interface{})}
	newNodeData.Set("nodeinfo.hostname", "Neuer nodename")

	res, err := merge(nodeData.Root, newNodeData.Root)
	assert.Nil(err)
	nodeData.Root = res
	newHostname, err := nodeData.Get("nodeinfo.hostname")
	assert.Nil(err)
	assert.Equal("Neuer nodename", newHostname.(string))
}

func TestTopLevelArrayMerge(t *testing.T) {
	assert := assert.New(t)

	dataParent := make([]interface{}, 3, 3)
	dataParent[0] = "item1"
	dataParent[1] = "item2"
	dataParent[2] = "item3"

	dataChild := make([]interface{}, 1, 1)
	dataChild[0] = "item3"

	res, err := merge(dataParent, dataChild)
	assert.Nil(err)
	assert.Equal(3, len(res.([]interface{})))

}
