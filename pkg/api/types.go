package api

type SmegNode struct {
	Alias       string            `json:"alias"`
	WgHost      string            `json:"wgHost"`
	WgEndpoint  string            `json:"wgEndpoint"`
	Endpoint    string            `json:"endpoint"`
	Timestamp   int               `json:"timestamp"`
	Description string            `json:"description"`
	PublicKey   string            `json:"publicKey"`
	Routes      []string          `json:"routes"`
	Services    map[string]string `json:"services"`
}

type SmegMesh struct {
	MeshId string              `json:"meshid"`
	Nodes  map[string]SmegNode `json:"nodes"`
}

type CreateMeshRequest struct {
	IfName string `json:"ifName" binding:"required"`
	WgPort int    `json:"port" binding:"required,gte=1024,lt=65535"`
}

type JoinMeshRequest struct {
	IfName    string `json:"ifName" binding:"required"`
	WgPort    int    `json:"port" binding:"required,gte=1024,lt=65535"`
	Bootstrap string `json:"bootstrap" binding:"required"`
	MeshId    string `json:"meshid" binding:"required"`
}
