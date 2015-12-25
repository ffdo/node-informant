package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergingDataIntoEmptyMap(t *testing.T) {
	assert := assert.New(t)

	collectedData := make(map[string]interface{})
	err := set(collectedData, "nodeinfo.hostname", "fftest")
	assert.Nil(err)
	assert.NotNil(collectedData["nodeinfo"])
	value, exists := collectedData["nodeinfo"]
	assert.True(exists)
	nodeMap, ok := value.(map[string]interface{})
	assert.True(ok)
	assert.Equal("fftest", nodeMap["hostname"])
}
