package uptime

import (
	"github.com/capnm/sysinfo"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector"
)

func Init() {
	uptimeCollectorCreator := collector.NewFactCollector("statistics.uptime", 60,
		initUptimeCollector, collectUptime)
	collector.Register(uptimeCollectorCreator)
}

type UptimeCollector struct{}

func initUptimeCollector(map[string]interface{}) error {
	// Just here for interface compliance
	return nil
}

func collectUptime() (interface{}, error) {
	return sysinfo.Get().Uptime.Seconds(), nil
}
