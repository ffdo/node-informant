package meshviewer

import (
	"testing"

	"encoding/json"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

func testForDoublettes(assert *assert.Assertions, nodes []*GraphNode) {
	for i, node := range nodes {
		for y, node2 := range nodes {
			if i != y {
				assert.NotEqual(node.NodeId, node2.NodeId, "Doublette node found at positions %d and %d", i, y)
			}
		}
	}
}

func TestGeneratingNodeGraph(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
	store := data.NewSimpleInMemoryStore()
	test.ExecuteCompletePipe(t, store)

	graphGenerator := &GraphGenerator{Store: store}
	graphGenerator.UpdateGraphJson()

	graphData := &GraphJson{}
	err := json.Unmarshal([]byte(graphGenerator.cachedJsonString), graphData)

	assert.Nil(err)
	assert.NotNil(graphData)
	testForDoublettes(assert, graphData.Batadv.Nodes)
}
