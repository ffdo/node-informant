package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestRegexp(t *testing.T) {
	assert := assert.New(t)

	assert.True(requestRegexp.MatchString("GET nodeinfo"))
	assert.False(requestRegexp.MatchString("GETIT nodeinfo"))
	assert.False(requestRegexp.MatchString("GET ../secretfile"))

	finds := requestRegexp.FindAllStringSubmatch("GET nodeinfo", -1)
	assert.Equal("nodeinfo", finds[0][1])
}
