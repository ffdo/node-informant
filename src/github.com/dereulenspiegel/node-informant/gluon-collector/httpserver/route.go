package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type HttpServeable interface {
	Routes() []Route
}

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
