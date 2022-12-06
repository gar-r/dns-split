package config

type Config struct {
	Addr   string        `json:"addr"`
	Dns    string        `json:"dns"`
	Tunnel *TunnelConfig `json:"split-tunnel"`
}

type TunnelConfig struct {
	Dns     string         `json:"dns"`
	Domains []string       `json:"domains"`
	Exclude []string       `json:"exclude"`
	Netlink *NetlinkConfig `json:"netlink"`
}

type NetlinkConfig struct {
	Enabled bool   `json:"enabled"`
	Dev     string `json:"dev"`
}
