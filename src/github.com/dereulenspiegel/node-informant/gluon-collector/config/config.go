package config

import (
	"flag"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/olebedev/config"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")

var Global *config.Config

func UInt(key string, def int) int {
	if Global != nil {
		return Global.UInt(key, def)
	} else {
		return def
	}
}

func UString(key, def string) string {
	if Global != nil {
		return Global.UString(key, def)
	} else {
		return def
	}
}

func ParseConfig(path string) error {
	var err error = nil
	if strings.HasSuffix(path, "yaml") {
		Global, err = config.ParseYamlFile(path)
	} else if strings.HasSuffix(path, "json") {
		Global, err = config.ParseJsonFile(path)
	}
	return err
}

func InitConfig() {
	err := ParseConfig(*configFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err,
			"configFilePath": *configFilePath,
		}).Warn("Unable to parse config file")
	}
}
