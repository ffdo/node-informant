package meshviewer

import (
	"testing"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
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
	assert.Equal(1, len(nodeList))
	assert.Equal("001122334455", nodeList[0].NodeId)
	// Access to the map after unmarshalling has no stable order, so both bat macs
	// could be in the id field, but either way should be fine
	assert.True(nodeList[0].Id == "001122aa44aa" || nodeList[0].Id == "001122aa44cc")
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
	log.SetLevel(log.ErrorLevel)
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
