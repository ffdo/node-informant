package frontend

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ffdo/node-informant/sensor/module"
	"github.com/olebedev/config"
)

type Frontend interface {
	io.Closer
	Start() error
}

type HttpFrontend interface {
	Frontend
	GetRoutes() []Route
}

type FrontendCreator func(*config.Config) (Frontend, error)

var (
	frontendCreators map[string]FrontendCreator
	frontends        []Frontend
)

func RegisterFrontendCreator(name string, creator FrontendCreator) {
	frontendCreators[name] = creator
}

func Init(cfg *config.Config) error {
	configureMartini()
	createRoutes()
	frontendConfigs, err := cfg.List("frontends")
	if err != nil {
		log.WithError(err).Fatal("No frontends configured")
	}
	for i, _ := range frontendConfigs {
		if frontendConfigs, err := cfg.Get(fmt.Sprintf("frontends.%d", i)); err != nil {
			log.WithError(err).Fatal("Error when retrieving frontend config")
		} else {
			if name, err := frontendConfigs.String("type"); err != nil {
				log.WithError(err).Fatal("Type for frontend is not specified")
			} else {
				creator := frontendCreators[name]
				if frontend, err := creator(frontendConfigs); err != nil {
					log.WithError(err).WithField("frontendType", name).Fatal("Can't create frontend")
				} else {
					frontends = append(frontends, frontend)
				}
			}
		}
	}
	return nil
}

func Start() error {
	for _, frontend := range frontends {
		if err := frontend.Start(); err != nil {
			return err
		}
		if httpFrontend, ok := frontend.(HttpFrontend); ok {
			log.Infof("Found http frontend %v", httpFrontend)
			addFrontendToHttp(httpFrontend)
		}
	}
	go http.ListenAndServe(":8088", m)
	return nil
}

func addFrontendToHttp(httpFrontent HttpFrontend) {
	for _, route := range httpFrontent.GetRoutes() {
		router.AddRoute(route.Method, route.Path, route.Handler)
	}
}

func Close() error {
	for _, frontend := range frontends {
		if err := frontend.Close(); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	frontendCreators = make(map[string]FrontendCreator)
	frontends = make([]Frontend, 0, 10)
	module.Register(module.Module{
		Init:  Init,
		Start: Start,
		Close: Close,
		Name:  "Frontend",
	})
}
