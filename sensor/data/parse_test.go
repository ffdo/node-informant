package data

import (
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
{"statistics":{"node_id":"60e327c7505a","clients":{"wifi":2,"total":2,"wifi24":2,"wifi5":0},"rootfs_usage":0.51785714285714275,"mesh_vpn":{"groups":{"do02":{"peers":{"do02100":null,"do02200":null}},"do01":{"peers":{"do01100":null,"do01200":{"established":414500.61300000006}}}}},"memory":{"cached":2620,"total":28540,"buffers":1140,"free":7044},"uptime":5133586.76,"idletime":4252001.5999999994,"gateway":"02:ce:ef:ca:fe:2b","processes":{"total":42,"running":1},"traffic":{"tx":{"packets":9491888,"dropped":25771,"bytes":1229916800},"rx":{"bytes":33359780712,"packets":324930660},"forward":{"bytes":28951610,"packets":131474},"mgmt_tx":{"bytes":38806569960.000002,"packets":110082408},"mgmt_rx":{"bytes":35525284330.000001,"packets":248869001}},"loadavg":0.18}}
`

func TestParsingStatistics(t *testing.T) {
	assert := assert.New(t)
	statistics, err := ParseJson([]byte(exampleStatistics))
	assert.Nil(err)
	assert.NotNil(statistics)
	gatewayInfo, err := statistics.Get("statistics.gateway")
	assert.Nil(err)
	assert.Equal("02:ce:ef:ca:fe:2b", gatewayInfo.(string))
}

func TestParsingNodeinfo(t *testing.T) {
	assert := assert.New(t)
	nodeData, err := ParseJson([]byte(exampleNodeInfo))
	assert.Nil(err)
	nodeId, err := nodeData.Get("nodeinfo.node_id")
	assert.Nil(err)
	assert.Equal("c46e1fb64f70", nodeId)
	autoupdaterEnabled, err := nodeData.Get("nodeinfo.software.autoupdater.enabled")
	assert.Nil(err)
	assert.True(autoupdaterEnabled.(bool))
}
