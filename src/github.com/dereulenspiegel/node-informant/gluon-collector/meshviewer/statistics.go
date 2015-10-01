package meshviewer

import "github.com/dereulenspiegel/node-informant/gluon-collector/data"

/*
"statistics": {
        "clients": 2,
        "gateway": "04ceefcafe2a",
        "loadavg": 0.56,
        "memory_usage": 0.26737595389633007,
        "rootfs_usage": 0.065068493150685,
        "traffic": {
          "forward": {
            "bytes": 2.398570215e+09,
            "packets": 7.704205e+06
          },
          "mgmt_rx": {
            "bytes": 1.5278906763e+10,
            "packets": 7.2848299e+07
          },
          "mgmt_tx": {
            "bytes": 9.80502624e+09,
            "packets": 2.8666078e+07
          },
          "rx": {
            "bytes": 3.937967996e+10,
            "packets": 6.7329241e+07
          },
          "tx": {
            "bytes": 2.88728849e+09,
            "dropped": 15309,
            "packets": 2.6166473e+07
          }
        },
        "uptime": 867590.55
      }
    }
*/

type StatisticsStruct struct {
	Clients     int                 `json:"clients"`
	Gateway     string              `json:"gateway"`
	Loadavg     float64             `json:"loadavg"`
	MemoryUsage float64             `json:"memory_usage"`
	RootfsUsage float64             `json:"rootfs_usage"`
	Uptime      float64             `json:"uptime"`
	Traffic     *data.TrafficStruct `json:"traffic"`
}
