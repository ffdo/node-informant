package meshviewer

import (
	"testing"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/stretchr/testify/assert"
)

func TestLoadingLegacyNodesJson(t *testing.T) {
	assert := assert.New(t)
	nodeJsonPath := "../../../../../../nodes.json"

	store := data.NewSimpleInMemoryStore()
	importer := &FFMapBackendDataLoader{Store: store}
	err := importer.LoadNodesFromFile(nodeJsonPath)
	assert.Nil(err)
	assert.True(0 < len(store.GetNodeInfos()))
}
