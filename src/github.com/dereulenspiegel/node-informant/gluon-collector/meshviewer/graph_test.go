package meshviewer

import (
	"testing"

	"encoding/json"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

var (
	neighbourInfos = []data.NeighbourStruct{
		data.NeighbourStruct{
			NodeId: "001122334455",
			Batadv: map[string]data.BatadvNeighbours{
				"001122aa44cc": data.BatadvNeighbours{
					Neighbours: map[string]data.BatmanLink{
						"ffbbaa33dd11": data.BatmanLink{
							Lastseen: 1.05,
							Tq:       233,
						},
					},
				},
				"001122aa44aa": data.BatadvNeighbours{
					Neighbours: map[string]data.BatmanLink{
						"ffbbaa33dd22": data.BatmanLink{
							Lastseen: 1.42,
							Tq:       55,
						},
					},
				},
			},
		},
	}

	nodeStatus = []data.NodeStatusInfo{
		data.NodeStatusInfo{
			NodeId: "001122334455",
			Online: true,
		},
	}
)

func TestCorrectGraphFormat(t *testing.T) {
	assert := assert.New(t)
	store := data.NewSimpleInMemoryStore()
	for _, neighbour := range neighbourInfos {
		store.PutNodeNeighbours(neighbour)
	}
	store.PutNodeStatusInfo("001122334455", nodeStatus[0])

	graphGenerator := &GraphGenerator{Store: store}
	_, nodeList := graphGenerator.buildNodeTableAndList()
	assert.NotNil(nodeList)
	assert.Equal(2, len(nodeList))

	assert.Equal("001122334455", nodeList[0].NodeId)
	assert.Equal("", nodeList[1].NodeId)
	assert.Equal("001122aa44cc", nodeList[0].Id)
	assert.Equal("001122aa44aa", nodeList[1].Id)
}

func testForDoublettes(assert *assert.Assertions, nodes []*GraphNode) {
	for i, node := range nodes {
		for y, node2 := range nodes {
			if i != y {
				assert.False(node.NodeId == node2.NodeId && node.Id == node2.Id,
					"Doublette node found at positions %d and %d with ids %s and %s", i, y, node.Id, node2.Id)
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
