package data

type TrafficObject struct {
	Bytes   uint64 `json:"bytes"`
	Packets uint64 `json:"packets"`
	Dropped uint64 `json:"dropped"`
}

type TrafficStruct struct {
	Tx      TrafficObject `json:"tx"`
	Rx      TrafficObject `json:"rx"`
	Forward TrafficObject `json:"forward"`
	MgmtTx  TrafficObject `json:"mgmt_tx"`
	MgmtRx  TrafficObject `json:"mgmt_rx"`
}

type StatisticsStruct struct {
	NodeId  string `json:"node_id"`
	Clients struct {
		Wifi  int `json:"wifi"`
		Total int `json:"total"`
	} `json:"clients"`
	RootFsUsage float64       `json:"rootfs_usage"`
	Traffic     TrafficStruct `json:"traffic"`
	Memory      struct {
		Cached  uint64 `json:"cached"`
		Total   uint64 `json:"total"`
		Buffers uint64 `json:"buffers"`
		Free    uint64 `json:"free"`
	} `json:"memory"`
	Uptime    float64 `json:"uptime"`
	Idletime  float64 `json:"idletime"`
	Gateway   string  `json:"gateway"`
	Processes struct {
		Total   uint64 `json:"total"`
		Running uint64 `json:"running"`
	} `json:"processes"`
	LoadAverage float64 `json:"loadavg"`
}
