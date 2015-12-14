package main

import (
	"net"
	"testing"
	"time"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

var (
	additionalData = []announced.Response{
		announced.Response{
			ClientAddr: &net.UDPAddr{
				IP:   net.ParseIP("fe80::1001:bcad"),
				Port: 4242,
			},
			Payload: []byte("Additional payload"),
		},
	}
)

func TestMultiReceiver(t *testing.T) {
	assert := assert.New(t)

	testReceiver1 := &TestDataReceiver{test.TestData}
	testReceiver2 := &TestDataReceiver{additionalData}

	multiReceiver := NewMultiReceiver(testReceiver1, testReceiver2)

	totalPacketCount := len(test.TestData) + len(additionalData)
	packetFound := false
	i := 0
	go multiReceiver.Receive(func(packet announced.Response) {
		i = i + 1
		payloadString := string(packet.Payload)
		if payloadString == "Additional payload" {
			packetFound = true
		}
	})
	testReceiver1.Close()
	testReceiver2.Close()
	for i < totalPacketCount {
		time.Sleep(time.Millisecond * 1)
	}
	assert.Equal(totalPacketCount, i, "Received less packets than we fed through 2 receiver")
	assert.True(packetFound, "Didn't found the additional payload")
}
