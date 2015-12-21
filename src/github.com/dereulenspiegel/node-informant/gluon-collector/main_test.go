package main

import (
	"os"
	"testing"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

func TestCompletePipe(t *testing.T) {
	store := data.NewSimpleInMemoryStore()
	test.ExecuteCompletePipe(t, store)
}

func TestCompletePipeWithBoltStore(t *testing.T) {
	assert := assert.New(t)
	dbPath := "./bolt.db"
	defer os.RemoveAll(dbPath)
	store, err := data.NewBoltStore(dbPath)
	assert.Nil(err)
	assert.NotNil(store)
	test.ExecuteCompletePipe(t, store)
	store.Close()
}
