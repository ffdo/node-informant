package data

type NetworkStruct struct {
	Mac       string   `json:"mac"`
	Addresses []string `json:"addresses"`
	Mesh      struct {
		Bat0 struct {
			Interfaces struct {
				Wireless []string `json:"wireless"`
				Other    []string `json:"other"`
				Tunnel   []string `json:"tunnel"`
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
	Longtitude float64 `json:"longtitude"`
	Latitude   float64 `json:"latitude"`
}

type SoftwareStruct struct {
	Autoupdater struct {
		Enabled bool   `json:"enabled"`
		Branch  string `json:"branch"`
	} `json:"autoupdater"`
	BatmanAdv struct {
		Version string `json:"version"`
		Compat  int    `json:"compat"`
	} `json:"batman-adv"`
	Fastd struct {
		Enabled bool   `json:"enabled"`
		Version string `json:"version"`
	} `json:"fastd"`
	Firmware struct {
		Base    string `json:"base"`
		Release string `json:"release"`
	} `json:"firmware"`
	StatusPage struct {
		Api int `json:"api"`
	} `json:"status-page"`
}

type HardwareStruct struct {
	Nproc int    `json:"nproc"`
	Model string `json:"model"`
}

type NodeInfo struct {
	NodeId   string         `json:"node_id"`
	Network  NetworkStruct  `json:"network"`
	Owner    OwnerStruct    `json:"owner"`
	System   SystemStruct   `json:"system"`
	Hostname string         `json:"hostname"`
	Location LocationStruct `json:"location"`
	Software SoftwareStruct `json:"software"`
	Hardware HardwareStruct `json:"hardware"`
}

type RespondNodeinfo struct {
	Nodeinfo   *NodeInfo         `json:"nodeinfo"`
	Statistics *StatisticsStruct `json:"statistics"`
	Neighbours *NeighbourStruct  `json:"neighbours"`
}
