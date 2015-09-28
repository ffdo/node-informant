package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleNodeInfo = `{"nodeinfo":
  {"node_id":"c46e1fb64f70",
   "network":
    {
     "mac":"c4:6e:1f:b6:4f:70",
     "addresses":["2a03:2260:50:5:c66e:1fff:feb6:4f70","fe80:0:0:0:c66e:1fff:feb6:4f70"],
     "mesh":
      {
        "bat0":
        {"interfaces":
          {"wireless":["c6:71:20:b6:4f:70"]}
        }
      },
      "mesh_interfaces":["c6:71:20:b6:4f:70"]
    },
    "owner": {"contact":"www.pauluskircheundkultur.net"},
    "system":
        {"site_code":"ffdo"},
        "hostname":"FF-DO-Paulus-Gemeindehaus02",
        "location":{"longitude":7.45485,"latitude":51.52076},
        "software":{
          "autoupdater":{
            "enabled":true,
            "branch":"stable"
          },
          "batman-adv":{
            "version":"2015.0",
            "compat":15
          },
          "fastd":{
            "enabled":false,
            "version":"v17"
          },
          "firmware":{
            "base":"gluon-v2015.1.2",
            "release":"0.7.2"
          }
        },
        "hardware":{
          "nproc":1,
          "model":
          "TP-Link TL-WR841N/ND v9"
        }
      }
    }`
var exampleStatistics = `
{
  "statistics":
    {
      "node_id":"c46e1fc70bbc",
      "clients":{
        "wifi":6,
        "total":6
      },
      "rootfs_usage":0.42361111111111,
      "traffic":{
        "tx":{
          "packets":5531082,
          "dropped":10562,
          "bytes":865662684
        },
        "rx":{
          "bytes":5690883474,
          "packets":33817859
        },
        "forward":{
          "bytes":324361367,
          "packets":859279
        },
        "mgmt_tx":{
          "bytes":2222548944,
          "packets":7179174
        },
        "mgmt_rx":{
          "bytes":2534888149,
          "packets":7531215
        }
      },
      "memory":{
        "cached":3124,
        "total":28860,
        "buffers":1548,
        "free":5356
      },
      "uptime":791551,
      "idletime":697373.8,
      "gateway":"02:ce:ef:ca:fe:2a",
      "processes":{
        "total":41,
        "running":1
      },
      "loadavg":0.01
    }
}
`

func TestParsingNodeinfo(t *testing.T) {
	assert := assert.New(t)
	nodeinfo := &RespondNodeinfo{}
	err := json.Unmarshal([]byte(exampleNodeInfo), nodeinfo)
	assert.Nil(err)
	assert.NotNil(nodeinfo.Nodeinfo)
	assert.Equal("c46e1fb64f70", nodeinfo.Nodeinfo.NodeId)
	assert.Equal("www.pauluskircheundkultur.net", nodeinfo.Nodeinfo.Owner.Contact)
}

func TestParsingNodeStatistics(t *testing.T) {
	assert := assert.New(t)
	nodeinfo := &RespondNodeinfo{}
	err := json.Unmarshal([]byte(exampleStatistics), nodeinfo)
	assert.Nil(err)
	assert.NotNil(nodeinfo.Statistics)
	assert.Equal("c46e1fc70bbc", nodeinfo.Statistics.NodeId)
	assert.Equal("02:ce:ef:ca:fe:2a", nodeinfo.Statistics.Gateway)
	assert.Equal(uint64(5531082), nodeinfo.Statistics.Traffic.Tx.Packets)
}
