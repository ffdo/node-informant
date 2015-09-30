package httpserver

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"

	log "github.com/Sirupsen/logrus"
	conf "github.com/dereulenspiegel/node-informant/gluon-collector/config"
)

func StartHttpServerBlocking(serveables ...HttpServeable) {
	router := AssembleRouter(serveables...)
	httpPort := conf.Global.UInt("http.port", 8080)
	httpAddress := conf.Global.UString("http.address", "")
	httpListenAddr := fmt.Sprintf("%s:%d", httpAddress, httpPort)
	log.Printf("Trying to http listen on %s", httpListenAddr)
	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(httpListenAddr, handler))
}
