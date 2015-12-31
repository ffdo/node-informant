package data

type NetworkStruct struct {
	Mac       string   `json:"mac"`
	Addresses []string `json:"addresses"`
	Mesh      struct {
		Bat0 struct {
			Interfaces struct {
				Wireless []string `json:"wireless,omitempty"`
				Other    []string `json:"other,omitempty"`
				Tunnel   []string `json:"tunnel,omitempty"`
			} `json:"interfaces"`
		} `json:"bat0"`
	} `json:"mesh"`
	MeshInterfaces []string `json:"mesh_interfaces"`
}

type OwnerStruct struct {
	Contact string `json:"contact"`
}

type SystemStruct struct {
	SiteCode string `json:"site_code"`
}

type LocationStruct struct {
	Longtitude float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Altitude   float64 `json:"altitude,omitempty"`
}

type SoftwareStruct struct {
	Autoupdater *struct {
		Enabled bool   `json:"enabled"`
		Branch  string `json:"branch"`
	} `json:"autoupdater,omitempty"`
	BatmanAdv *struct {
		Version string `json:"version"`
		Compat  int    `json:"compat"`
	} `json:"batman-adv,omitempty"`
	Fastd *struct {
		Enabled bool   `json:"enabled"`
		Version string `json:"version"`
	} `json:"fastd,omitempty"`
	Firmware *struct {
		Base    string `json:"base"`
		Release string `json:"release"`
	} `json:"firmware,omitempty"`
	StatusPage *struct {
		Api int `json:"api"`
	} `json:"status-page,omitempty"`
}

type HardwareStruct struct {
	Nproc int    `json:"nproc"`
	Model string `json:"model"`
}

type NodeInfo struct {
	NodeId   string          `json:"node_id"`
	Network  NetworkStruct   `json:"network"`
	Owner    *OwnerStruct    `json:"owner,omitempty"`
	System   SystemStruct    `json:"system"`
	Hostname string          `json:"hostname"`
	Location *LocationStruct `json:"location,omitempty"`
	Software SoftwareStruct  `json:"software"`
	Hardware HardwareStruct  `json:"hardware"`
}

type RespondNodeinfo struct {
	Nodeinfo   *NodeInfo         `json:"nodeinfo"`
	Statistics *StatisticsStruct `json:"statistics"`
	Neighbours *NeighbourStruct  `json:"neighbours"`
}
