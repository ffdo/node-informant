package module

import (
	log "github.com/Sirupsen/logrus"
	"github.com/olebedev/config"
)

var (
	registeredModules []Module
)

func init() {
	registeredModules = make([]Module, 0, 10)
}

func Register(mod Module) {
	registeredModules = append(registeredModules, mod)
}

func InitAllModules(cfg *config.Config) error {
	for _, registeredModule := range registeredModules {
		log.WithField("moduleName", registeredModule.Name).Debug("Initialising module")
		if err := registeredModule.Init(cfg); err != nil {
			return err
		}
	}
	return nil
}

func StartAllModules() error {
	for _, registeredModule := range registeredModules {
		log.WithField("moduleName", registeredModule.Name).Debug("Starting module")
		if err := registeredModule.Start(); err != nil {
			log.WithError(err).WithField("moduleName", registeredModule.Name).Error("Failed to start module")
			return err
		}
	}
	return nil
}

func CloseAllModules(exitOnFail bool) error {
	for _, registeredModule := range registeredModules {
		if err := registeredModule.Close(); err != nil && exitOnFail {
			return err
		} else if err != nil {
			log.WithError(err).Error("Can't close module properly")
		}
	}
	return nil
}

type Module struct {
	Init  InitModule
	Start StartModule
	Close CloseModule
	Name  string
}

type InitModule func(*config.Config) error
type StartModule func() error
type CloseModule func() error
