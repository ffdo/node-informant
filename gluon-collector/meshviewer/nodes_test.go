package meshviewer

import (
	"encoding/json"
	"testing"

	"github.com/ffdo/node-informant/gluon-collector/data"
	"github.com/ffdo/node-informant/gluon-collector/test"
	"github.com/stretchr/testify/assert"
)

var (
	nodeinfos = `
	{
  "hardware": {
    "model": "TP-Link TL-WR842N/ND v2",
    "nproc": 1
  },
  "hostname": "FF-DO-Josephstr-13",
  "location": {
    "latitude": 51.512426,
    "longitude": 7.455485
  },
  "network": {
    "addresses": [
      "fe80:0:0:0:eade:27ff:fe25:2554",
      "2a03:2260:50:5:eade:27ff:fe25:2554"
    ],
    "mac": "e8:de:27:25:25:54",
    "mesh": {
      "bat0": {
        "interfaces": {
          "tunnel": [
            "ea:e2:27:25:25:54"
          ],
          "wireless": [
            "ea:e1:28:25:25:54"
          ]
        }
      }
    },
    "mesh_interfaces": [
      "ea:e2:27:25:25:54",
      "ea:e1:28:25:25:54"
    ]
  },
  "node_id": "e8de27252554",
  "owner": {
    "contact": "till.klocke@gmail.com"
  },
  "software": {
    "autoupdater": {
      "branch": "stable",
      "enabled": true
    },
    "batman-adv": {
      "compat": 15,
      "version": "2015.0"
    },
    "fastd": {
      "enabled": true,
      "version": "v17"
    },
    "firmware": {
      "base": "gluon-v2015.1.2",
      "release": "0.7.2"
    }
  },
  "system": {
    "site_code": "ffdo"
  }
}
	`
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

func TestNodesJsonHasFlaggedUplinks(t *testing.T) {
	assert := assert.New(t)
	store := data.NewSimpleInMemoryStore()
	test.ExecuteCompletePipe(t, store)

	statistics := data.StatisticsStruct{}
	err := json.Unmarshal([]byte(uplinkStats), &statistics)
	assert.Nil(err)

	store.PutStatistics(statistics)

	generator := &NodesJsonGenerator{Store: store}
	nodesJson := generator.GetNodesJson()

	uplinkFound := false
	for _, node := range nodesJson.Nodes {
		if node.Flags.Uplink && node.Nodeinfo.NodeId == "e8de27252554" {
			uplinkFound = true
			break
		}
	}
	assert.True(uplinkFound, "No node was flagged as uplink")

	statistics = data.StatisticsStruct{}
	err = json.Unmarshal([]byte(nonUplinkStats), &statistics)
	assert.Nil(err)

	uplinkFound = false
	store.PutStatistics(statistics)
	nodesJson = generator.GetNodesJson()
	for _, node := range nodesJson.Nodes {
		if !node.Flags.Uplink && node.Nodeinfo.NodeId == "e8de27252554" {
			uplinkFound = true
			break
		}
	}
	assert.True(uplinkFound, "A node was not correctly flagged")
}

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
