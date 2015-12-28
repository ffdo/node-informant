package nodeid

import (
	"fmt"
	"net"
	"strings"

	"github.com/dereulenspiegel/node-informant/gluon-emitter/collector"
	"github.com/dereulenspiegel/node-informant/gluon-emitter/data"
)

func Init() {
	nodeidCollector := collector.NewFactCollector("nodeinfo.node_id", collector.DontPoll,
		initNodeidCollector, collectNodeid)
	collector.Register(nodeidCollector)
}

func initNodeidCollector(config map[string]interface{}) error {
	if value, exists := config["interface"]; exists {
		interfaceName := value.(string)
		iface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			return err
		}
		mac := iface.HardwareAddr.String()
		nodeid := strings.Replace(mac, ":", "", -1)
		data.MergeCollectedData("statistics.node_id", nodeid)
		data.MergeCollectedData("nodeinfo.node_id", nodeid)
		data.MergeCollectedData("neighbours.node_id", nodeid)

		return nil
	} else {
		return fmt.Errorf("No interface defined")
	}
}

func collectNodeid() (interface{}, error) {
	return nil, fmt.Errorf("Shouldn't happen")
}
