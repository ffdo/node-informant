package hostname

import (
	"os"

	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector"
)

func createHostNameCollector() collector.FactCollector {
	return &HostnameCollector{}
}

func Init() {
	collector.Register(createHostNameCollector)
}

type HostnameCollector struct{}

func (h *HostnameCollector) Init(map[string]interface{}) error {
	// Just here for interface compliance
	return nil
}

func (h *HostnameCollector) Collect() (interface{}, error) {
	return os.Hostname()
}

func (h *HostnameCollector) Path() string {
	return "nodeinfo.system.hostname"
}

func (h *HostnameCollector) Frequency() collector.CollectionFrequency {
	return collector.CollectOnce
}
