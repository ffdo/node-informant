package data

import (
	"encoding/json"
	"testing"

	"github.com/dereulenspiegel/node-informant/utils"
	"github.com/stretchr/testify/assert"
)

var (
	expectedJson = `{"interfaces":{"mesh":"interface1"},"node_id":"test"}`

	badMap = map[interface{}]interface{}{
		"key1": map[interface{}]interface{}{
			"nestedKey1": "nestedData",
		},
		"key2": "unnestedData",
	}
)

func TestNormalization(t *testing.T) {
	assert := assert.New(t)

	saneMap, err := normalize(badMap)
	assert.Nil(err)
	assert.NotNil(saneMap)
	_, ok := saneMap.(map[string]interface{})
	assert.True(ok)

	//assert.Equal("nestedData", stringMap["key1"]["nestedKey1"])
	//assert.Equal("unnestedData", stringMap["key2"])

	out, err := json.Marshal(&saneMap)
	assert.Nil(err)
	assert.NotNil(out)
	assert.True(len(out) > 0)
}

func TestRetrievalOfData(t *testing.T) {
	assert := assert.New(t)
	nodeinfos := map[string]interface{}{
		"node_id": "test",
		"interfaces": map[string]interface{}{
			"mesh": "interface1",
		},
	}
	collectedData["nodeinfo"] = nodeinfos

	out, err := GetMarshalledAndCompressedSection("nodeinfo")
	assert.Nil(err)
	assert.True(len(out) > 0)

	uncompressed, err := utils.Deflate(out)
	assert.Nil(err)
	assert.Equal(expectedJson, string(uncompressed))
}
