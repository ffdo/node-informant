package rest

import (
	"encoding/json"
	"net/http"

	"github.com/ffdo/node-informant/sensor/frontend"
	"github.com/ffdo/node-informant/sensor/store"
	"github.com/olebedev/config"
)

type RestFrontend struct{}

func init() {
	frontend.RegisterFrontendCreator("rest", createRestFrontend)
}

func createRestFrontend(cfg *config.Config) (frontend.Frontend, error) {
	return &RestFrontend{}, nil
}

func (r *RestFrontend) Close() error {
	return nil
}

func (r *RestFrontend) Start() error {
	return nil
}

func (r *RestFrontend) GetRoutes() []frontend.Route {
	return []frontend.Route{
		frontend.Route{
			Handler: GetAllNodeInfos,
			Method:  "GET",
			Path:    "/nodeinfos",
		},
	}
}

func GetAllNodeInfos(db store.Storage) (int, string) {
	allNodes := db.GetAllNodeData()
	return http.StatusOK, mustEncode(allNodes)
}

func mustEncode(data interface{}) string {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(dataBytes)
}
