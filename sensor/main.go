package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	_ "github.com/ffdo/node-informant/sensor/anncd"
	"github.com/ffdo/node-informant/sensor/module"
	"github.com/ffdo/node-informant/sensor/process"
	_ "github.com/ffdo/node-informant/sensor/rest"
	"github.com/ffdo/node-informant/sensor/store"
	_ "github.com/ffdo/node-informant/sensor/store/memory"
	"github.com/olebedev/config"
)

var (
	configPath = flag.String("config", "", "Path to the configuration file")
)

func configureLogger(cfg *config.Config) {
	log.SetLevel(log.DebugLevel)
}

func main() {
	flag.Parse()
	cfg, err := config.ParseYamlFile(*configPath)
	if err != nil {
		log.WithError(err).WithField("configPath", *configPath).Fatal("Can't parse config")
	}
	configureLogger(cfg)

	if storConfig, err := cfg.Get("storage"); err != nil {
		log.WithError(err).Fatal("Storage configuration doesn't exist")
	} else if err := store.ConfigureStorage(storConfig); err != nil {
		log.WithError(err).Fatal("Can't initialise storage")
	}

	if err := module.InitAllModules(cfg); err != nil {
		log.WithError(err).Fatal("Error while initialising the modules")
	}

	if err := module.StartAllModules(); err != nil {
		log.WithError(err).Fatal("Error while starting all modules")
	}

	process.ProcessWaitGroup.Wait()
	module.CloseAllModules(false)
}
