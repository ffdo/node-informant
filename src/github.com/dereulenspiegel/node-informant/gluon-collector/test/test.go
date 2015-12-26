package test

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	cfg "github.com/olebedev/config"
	"github.com/stretchr/testify/assert"

	"github.com/dereulenspiegel/node-informant/announced"
	"github.com/dereulenspiegel/node-informant/gluon-collector/assemble"
	"github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/dereulenspiegel/node-informant/gluon-collector/prometheus"
	"github.com/dereulenspiegel/node-informant/utils"
)

var TestData []announced.Response

type TestDataReceiver struct {
	TestData []announced.Response
}

func (t *TestDataReceiver) Query(queryString string) {
	//Nothing just here for interface compatibility
}

func (t *TestDataReceiver) QueryUnicast(addr *net.UDPAddr, queryString string) {
	//Nothing just here for interface compatibility
}

func (t *TestDataReceiver) Receive(rFunc func(announced.Response)) {
	for _, data := range t.TestData {
		rFunc(data)
	}
}

func (t *TestDataReceiver) Close() error {
	// Only here to satisfy the announced.AnnouncedPacketReceiver interface
	return nil
}

func ExecuteCompletePipe(t *testing.T, store data.Nodeinfostore) {
	prometheus.Init()
	assert := assert.New(t)
	testReceiver := &TestDataReceiver{TestData: TestData}

	i := 0
	closeables, err := assemble.BuildPipelines(store, testReceiver, func(response data.ParsedResponse) {
		i = i + 1
	})
	assert.Nil(err)

	for i < len(TestData) {
		time.Sleep(time.Millisecond * 1)
	}

	for _, closable := range closeables {
		closable.Close()
	}

	assert.Equal(len(TestData), i)
}

func stringToBytes(in string) []byte {
	parts := strings.Split(in, " ")
	bytes := make([]byte, len(parts))
	if len(parts) == 0 {
		log.Fatalf("No bytes to decode")
	}
	for i, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}
		y, err := strconv.Atoi(p)
		if err != nil {
			log.Fatalf("Can't convert string %s to int", p)
		}
		bytes[i] = byte(y)
	}
	return bytes
}

func findTestData() string {
	dataFound := false
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error can't determine current work directory: %v", err)
	}
	currentPath = path.Join(currentPath, "..")
	currentPath = path.Clean(currentPath)
	for !dataFound && currentPath != "/" {
		dataPath := path.Join(currentPath, "testdata.raw")
		if utils.FileExists(dataPath) {
			return dataPath
		}
		currentPath = path.Join(currentPath, "..")
		currentPath = path.Clean(currentPath)
	}
	return ""
}

func LoadTestData() error {
	dataPath := findTestData()
	//dataFile, err := os.Open("../../../../../../testdata.raw")
	dataFile, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	dataBytes, err := ioutil.ReadAll(dataFile)
	if err != nil {
		return err
	}
	dataString := string(dataBytes)
	serializedPacktes := strings.Split(dataString, "|")
	responses := make([]announced.Response, 0, len(serializedPacktes))
	for _, text := range serializedPacktes {
		if strings.TrimSpace(text) == "" {
			continue
		}
		printableResponse := &announced.PrintableResponse{}
		err := json.Unmarshal([]byte(text), printableResponse)
		if err != nil {
			log.Printf("Error unmarshalling json %s: %v", text, err)
			return err
		}
		addr := &net.UDPAddr{
			IP:   net.ParseIP(printableResponse.Addr.IP),
			Port: printableResponse.Addr.Port,
			Zone: printableResponse.Addr.Zone,
		}
		response := announced.Response{
			ClientAddr: addr,
			Payload:    stringToBytes(printableResponse.Bytes),
		}
		responses = append(responses, response)
	}
	TestData = responses
	log.Printf("Loaded %d packets", len(TestData))
	return nil
}

func initDefaultConfig() {
	config.Global = &cfg.Config{}
}

func init() {
	initDefaultConfig()
	err := LoadTestData()
	if err != nil {
		log.Fatalf("Can't load test data: %v", err)
	}
}
