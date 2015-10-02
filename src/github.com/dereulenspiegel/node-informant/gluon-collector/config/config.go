package config

import (
	"flag"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/olebedev/config"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")

var Global *config.Config

func ParseConfig(path string) error {
	var err error = nil
	if strings.HasSuffix(path, "yaml") {
		Global, err = config.ParseYamlFile(path)
	} else if strings.HasSuffix(path, "json") {
		Global, err = config.ParseJsonFile(path)
	}
	if err != nil {
		Global = &config.Config{}
	}
	return err
}

func init() {
	flag.Parse()
	err := ParseConfig(*configFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err,
			"configFilePath": *configFilePath,
		}).Fatal("Unable to parse config file, falling back to default values")
	}
}
