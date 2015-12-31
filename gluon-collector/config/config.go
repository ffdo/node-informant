package config

import (
	"flag"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/olebedev/config"
)

var configFilePath = flag.String("config", "/etc/node-collector.yaml", "Config file")

// Global represents the global configuration state.
var Global *config.Config

// UInt tries to retrieve an integer value specified by the key. If the global
// configuration state is null it returns the default. If the global configuration
// state is not null, the global configuration still may return the default value.
func UInt(key string, def int) int {
	if Global != nil {
		return Global.UInt(key, def)
	} else {
		return def
	}
}

// UString tries ro etrieve a string value specified by the key. For this same rules
// as with UInt apply.
func UString(key, def string) string {
	if Global != nil {
		return Global.UString(key, def)
	} else {
		return def
	}
}

// ParseConfig parses a configuration file located at path. Configuration files
// need to end in .yaml or .json, depending whether they are in yaml or json format
func ParseConfig(path string) error {
	var err error = nil
	if strings.HasSuffix(path, "yaml") {
		Global, err = config.ParseYamlFile(path)
	} else if strings.HasSuffix(path, "json") {
		Global, err = config.ParseJsonFile(path)
	}
	return err
}

// InitConfig parses the configuration located at the path specified via the config
// flag. This method should be called rather early in main.
func InitConfig() {
	err := ParseConfig(*configFilePath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err,
			"configFilePath": *configFilePath,
		}).Warn("Unable to parse config file")
	}
}
