package data

import (
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestStoringNodeinfo(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	assert := assert.New(t)

	dbPath := "./bolt.db"
	defer os.RemoveAll(dbPath)

	store, err := NewBoltStore(dbPath)
	assert.Nil(err)
	assert.NotNil(store)
	nodeinfo := NodeInfo{
		NodeId: "a",
	}
	store.PutNodeInfo(nodeinfo)
	result, err := store.GetNodeInfo("a")
	assert.Nil(err)
	assert.Equal("a", result.NodeId)
	store.Close()
}
