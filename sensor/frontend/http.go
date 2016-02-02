package frontend

import (
	"github.com/ffdo/node-informant/sensor/store"
	"github.com/go-martini/martini"
)

var (
	m      *martini.Martini
	router martini.Router
)

type Route struct {
	Path    string
	Method  string
	Handler martini.Handler
}

func configureMartini() {
	m = martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.MapTo(store.DB, (*store.Storage)(nil))
}

func createRoutes() {
	router = martini.NewRouter()
	m.MapTo(router, (*martini.Routes)(nil))
	m.Action(router.Handle)
}

func InitRestAPI() {
	configureMartini()
	createRoutes()
	m.Run()
}
