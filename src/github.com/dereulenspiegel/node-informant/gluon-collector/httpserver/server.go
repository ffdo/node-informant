package httpserver

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func StartHttpServerBlocking(httpListenAddr string, serveables ...HttpServeable) {
	router := AssembleRouter(serveables...)
	log.Printf("Trying to http listen on %s", httpListenAddr)
	log.Fatal(http.ListenAndServe(httpListenAddr, router))
}
