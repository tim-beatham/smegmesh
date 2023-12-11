package api

import "time"

type Route struct {
	Prefix string   `json:"prefix"`
	Path   []string `json:"path"`
}

type SmegStats struct {
	TotalTransmit     int64         `json:"totalTransmit"`
	TotalReceived     int64         `json:"totalReceived"`
	KeepAliveInterval time.Duration `json:"keepaliveInterval"`
	AllowedIps        []string      `json:"allowedIps"`
}

type SmegNode struct {
	Alias       string            `json:"alias"`
	WgHost      string            `json:"wgHost"`
	WgEndpoint  string            `json:"wgEndpoint"`
	Endpoint    string            `json:"endpoint"`
	Timestamp   int               `json:"timestamp"`
	Description string            `json:"description"`
	PublicKey   string            `json:"publicKey"`
	Routes      []Route           `json:"routes"`
	Services    map[string]string `json:"services"`
	Stats       SmegStats         `json:"stats"`
}

type SmegMesh struct {
	MeshId string              `json:"meshid"`
	Nodes  map[string]SmegNode `json:"nodes"`
}

type CreateMeshRequest struct {
	WgPort int `json:"port" binding:"omitempty,gte=1024,lt=65535"`
}

type JoinMeshRequest struct {
	WgPort    int    `json:"port" binding:"omitempty,gte=1024,lt=65535"`
	Bootstrap string `json:"bootstrap" binding:"required"`
	MeshId    string `json:"meshid" binding:"required"`
}

type ApiServerConf struct {
	WordsFile string
}
