package data

type MeshVPNPeerLink struct {
	Established float64 `json:"established"`
}

type MeshVPNPeerGroup struct {
	Peers  map[string]*MeshVPNPeerLink  `json:"peers"`
	Groups map[string]*MeshVPNPeerGroup `json:"groups"`
}

type MeshVPNStruct struct {
	Groups map[string]*MeshVPNPeerGroup `json:"groups,omitempty"`
}

type TrafficObject struct {
	Bytes   float64 `json:"bytes,omitempty"`
	Packets uint64  `json:"packets,omitempty"`
	Dropped uint64  `json:"dropped,omitempty"`
}

type TrafficStruct struct {
	Tx      *TrafficObject `json:"tx"`
	Rx      *TrafficObject `json:"rx"`
	Forward *TrafficObject `json:"forward"`
	MgmtTx  *TrafficObject `json:"mgmt_tx"`
	MgmtRx  *TrafficObject `json:"mgmt_rx"`
}

type ClientStatistics struct {
	Wifi  int `json:"wifi"`
	Total int `json:"total"`
}

type MemoryStatistics struct {
	Cached  uint64 `json:"cached"`
	Total   uint64 `json:"total"`
	Buffers uint64 `json:"buffers"`
	Free    uint64 `json:"free"`
}

type StatisticsStruct struct {
	NodeId      string           `json:"node_id"`
	Clients     ClientStatistics `json:"clients"`
	RootFsUsage float64          `json:"rootfs_usage"`
	Traffic     *TrafficStruct   `json:"traffic"`
	Memory      MemoryStatistics `json:"memory"`
	Uptime      float64          `json:"uptime"`
	Idletime    float64          `json:"idletime"`
	Gateway     string           `json:"gateway"`
	Processes   struct {
		Total   uint64 `json:"total"`
		Running uint64 `json:"running"`
	} `json:"processes"`
	LoadAverage float64        `json:"loadavg"`
	MeshVpn     *MeshVPNStruct `json:"mesh_vpn,omitempty"`
}
