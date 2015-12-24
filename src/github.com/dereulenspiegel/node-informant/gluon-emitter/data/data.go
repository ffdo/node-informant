package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/dereulenspiegel/node-informant/utils"
	"gopkg.in/yaml.v2"
)

var (
	collectedData map[string]interface{}
)

func init() {
	collectedData = make(map[string]interface{})
}

type configVars struct {
	Env map[string]string
}

func buildConfigVars() configVars {
	vars := configVars{
		Env: make(map[string]string),
	}
	for _, envVar := range os.Environ() {
		parts := strings.Split(envVar, "=")
		if len(parts) == 2 {
			vars.Env[parts[0]] = parts[1]
		}
	}
	return vars
}

func LoadYamlFile(path string) error {
	aliasTemplate := template.New("alias")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	aliasTemplate, err = aliasTemplate.Parse(string(data))
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	err = aliasTemplate.Execute(buffer, buildConfigVars())
	if err != nil {
		return err
	}

	alias := make(map[string]interface{})
	err = yaml.Unmarshal(buffer.Bytes(), &alias)
	if err != nil {
		return err
	}
	collectedData = alias
	return nil
}

func MergeCollectedData(path string, metrics map[string]interface{}) {
	if err := set(collectedData, path, metrics); err != nil {
		log.WithFields(log.Fields{
			"path":          path,
			"value":         metrics,
			"collectedData": collectedData,
		}).Errorf("Can't merge collected data")
	}
}

func GetMarshalledAndCompressedSection(section string) ([]byte, error) {
	dataSection, exists := collectedData[section]
	if !exists {
		return nil, fmt.Errorf("Section %s does not exist", section)
	}
	jsonData, err := json.Marshal(dataSection)
	if err != nil {
		return nil, err
	}
	compressedData, err := utils.DeflateCompress(jsonData)
	if err != nil {
		return nil, err
	}
	return compressedData, nil
}
