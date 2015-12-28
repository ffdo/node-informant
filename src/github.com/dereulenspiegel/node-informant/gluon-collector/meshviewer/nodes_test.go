package meshviewer

import (
	"encoding/json"
	"testing"

	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
	"github.com/stretchr/testify/assert"
)

var (
	uplinkStats = `
  {
        "mesh_vpn": {
          "groups": {
            "backbone": {
              "peers": {
                "w0": null,
                "w1": null,
                "w2": null,
                "w3": null,
                "w4": {
                  "established": 403368.478
                },
                "w5": null,
                "w6": {
                  "established": 433269.433
                },
                "w7": null,
                "w8": null,
                "w9": null
              }
            }
          }
        },
        "clients": {
    "total": 2,
    "wifi": 2
  },
  "gateway": "02:ce:ef:ca:fe:2a",
  "idletime": 418509.87,
  "loadavg": 0.1,
  "memory": {
    "buffers": 1516,
    "cached": 5288,
    "free": 3792,
    "total": 28860
  },
  "node_id": "e8de27252554",
  "processes": {
    "running": 1,
    "total": 41
  },
  "rootfs_usage": 0.066780821917808,
  "traffic": {
    "forward": {
      "bytes": 434509,
      "packets": 2260
    },
    "mgmt_rx": {
      "bytes": 200750066,
      "packets": 16514592
    },
    "mgmt_tx": {
      "bytes": 3673142686,
      "packets": 10716657
    },
    "rx": {
      "bytes": 2593562021,
      "packets": 24162988
    },
    "tx": {
      "bytes": 498811113,
      "dropped": 2152,
      "packets": 2318108
    }
  },
  "uptime": 494187.22
}
  `

	nonUplinkStats = `
  {
        "mesh_vpn": {
          "groups": {
            "backbone": {
              "peers": {
                "w0": null,
                "w1": null,
                "w2": null,
                "w3": null,
                "w4": null,
                "w5": null,
                "w6": null,
                "w7": null,
                "w8": null,
                "w9": null
              }
            }
          }
        },
        "clients": {
    "total": 2,
    "wifi": 2
  },
  "gateway": "02:ce:ef:ca:fe:2a",
  "idletime": 418509.87,
  "loadavg": 0.1,
  "memory": {
    "buffers": 1516,
    "cached": 5288,
    "free": 3792,
    "total": 28860
  },
  "node_id": "e8de27252554",
  "processes": {
    "running": 1,
    "total": 41
  },
  "rootfs_usage": 0.066780821917808,
  "traffic": {
    "forward": {
      "bytes": 434509,
      "packets": 2260
    },
    "mgmt_rx": {
      "bytes": 200750066,
      "packets": 16514592
    },
    "mgmt_tx": {
      "bytes": 3673142686,
      "packets": 10716657
    },
    "rx": {
      "bytes": 2593562021,
      "packets": 24162988
    },
    "tx": {
      "bytes": 498811113,
      "dropped": 2152,
      "packets": 2318108
    }
  },
  "uptime": 494187.22
  }
  `
)

func TestFlaggingUplink(t *testing.T) {
	assert := assert.New(t)
	statistics := data.StatisticsStruct{}
	err := json.Unmarshal([]byte(uplinkStats), &statistics)
	assert.Nil(err)
	assert.True(determineUplink(statistics))

	err = json.Unmarshal([]byte(nonUplinkStats), &statistics)
	assert.Nil(err)
	assert.False(determineUplink(statistics))
}
