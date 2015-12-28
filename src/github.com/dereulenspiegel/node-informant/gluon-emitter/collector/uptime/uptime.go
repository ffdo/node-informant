package uptime

import (
	"github.com/capnm/sysinfo"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector"
)

func Init() {
	uptimeCollectorCreator := collector.NewFactCollector("statistics.uptime", 60,
		nil, collectUptime)
	collector.Register(uptimeCollectorCreator)
}

func collectUptime() (interface{}, error) {
	return sysinfo.Get().Uptime.Seconds(), nil
}
