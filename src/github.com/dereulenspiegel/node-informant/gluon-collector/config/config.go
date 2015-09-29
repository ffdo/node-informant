package config

import (
	"strings"

	"github.com/olebedev/config"
)

var Global *config.Config

func ParseConfig(path string) error {
	var err error = nil
	if strings.HasSuffix(path, "yaml") {
		Global, err = config.ParseYamlFile(path)
	} else if strings.HasSuffix(path, "json") {
		Global, err = config.ParseJsonFile(path)
	}

	return err
}
