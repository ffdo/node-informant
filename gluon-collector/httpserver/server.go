package httpserver

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"

	stat "github.com/prometheus/client_golang/prometheus"

	log "github.com/Sirupsen/logrus"
	conf "github.com/ffdo/node-informant/gluon-collector/config"
)

// StartHttpServerBlocking is a blocking method which takes an arbritrary number
// of HttpServeables, adds all their routes the http mux and starts a server on
// the configured port and address.
func StartHttpServerBlocking(serveables ...HttpServeable) {
	router := AssembleRouter(serveables...)
	httpPort := conf.Global.UInt("http.port", 8080)
	httpAddress := conf.Global.UString("http.address", "")
	httpListenAddr := fmt.Sprintf("%s:%d", httpAddress, httpPort)
	log.Printf("Trying to http listen on %s", httpListenAddr)
	handler := cors.Default().Handler(router)
	http.Handle("/", handler)
	http.Handle("/metrics", stat.Handler())
	log.Fatal(http.ListenAndServe(httpListenAddr, nil))
}
