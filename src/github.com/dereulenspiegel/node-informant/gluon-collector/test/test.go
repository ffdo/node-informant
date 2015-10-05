package test

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/dereulenspiegel/node-informant/announced"
)

var TestData []announced.Response

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

func LoadTestData() error {
	dataFile, err := os.Open("../../../../../testdata.raw")
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

func init() {
	err := LoadTestData()
	if err != nil {
		log.Fatalf("Can't load test data: %v", err)
	}
}
