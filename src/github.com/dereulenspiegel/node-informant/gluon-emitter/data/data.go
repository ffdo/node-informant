package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func LoadAliases(filePath string) error {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".yml":
		return LoadYamlFile(filePath)
	case ".json":
		return LoadJsonFile(filePath)
	default:
		return fmt.Errorf("Unknown file format %s", ext)
	}
}

func templateAliasesFile(filePath string) ([]byte, error) {
	aliasTemplate := template.New("alias")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	aliasTemplate, err = aliasTemplate.Parse(string(data))
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	err = aliasTemplate.Execute(buffer, buildConfigVars())
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func LoadJsonFile(path string) error {
	data, err := templateAliasesFile(path)
	if err != nil {
		return err
	}
	alias := make(map[string]interface{})
	err = json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	collectedData = alias
	return nil
}

func LoadYamlFile(path string) error {
	data, err := templateAliasesFile(path)
	if err != nil {
		return err
	}

	alias := make(map[string]interface{})
	err = yaml.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	normalizedData, err := normalize(alias)
	collectedData = normalizedData.(map[string]interface{})
	return err
}

func normalize(in interface{}) (interface{}, error) {
	switch in.(type) {
	case map[string]interface{}:
		inMap := in.(map[string]interface{})
		for key, value := range inMap {
			var err error
			inMap[key], err = normalize(value)
			if err != nil {
				return nil, err
			}
		}
		return inMap, nil

	case map[interface{}]interface{}:
		stringMap := make(map[string]interface{})
		for key, value := range in.(map[interface{}]interface{}) {
			if stringKey, ok := key.(string); ok {
				var err error
				stringMap[stringKey], err = normalize(value)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("Can't convert %v to string", key)
			}
		}
		return stringMap, nil
	default:
		return in, nil
	}
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
	log.Debugf("Found data for section %s: %v", section, dataSection)
	jsonData, err := json.Marshal(&dataSection)
	if err != nil {
		return nil, err
	}
	log.Debugf("Marshalled data into json: %s", string(jsonData))
	compressedData, err := utils.DeflateCompress(jsonData)
	if err != nil {
		return nil, err
	}
	log.Debugf("Compressed data to %d bytes", len(compressedData))
	return compressedData, nil
}
