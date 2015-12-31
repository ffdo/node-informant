package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Reoute is struct representing a http route to be created by a http mux.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// HttpServeable is an interface which should be implemented by everything which
// wants to expose an API via http. Every Route returned in the slice will be added
// to the servers http mux.
type HttpServeable interface {
	Routes() []Route
}

// AssembleRouter takes an arbritrary count of HttpServeables and adds all their
// routes to its http mux.
func AssembleRouter(serveables ...HttpServeable) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, serveable := range serveables {
		for _, route := range serveable.Routes() {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(route.HandlerFunc)
		}
	}
	return router
}
