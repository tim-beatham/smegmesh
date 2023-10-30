package crdt

// MeshNodeCrdt: Represents a CRDT for a mesh nodes
type MeshNodeCrdt struct {
	HostEndpoint string                 `automerge:"hostEndpoint" json:"hostEndpoint"`
	WgEndpoint   string                 `automerge:"wgEndpoint" json:"wgEndpoint"`
	PublicKey    string                 `automerge:"publicKey" json:"publicKey"`
	WgHost       string                 `automerge:"wgHost" json:"wgHost"`
	Timestamp    int64                  `automerge:"timestamp" json:"timestamp"`
	Routes       map[string]interface{} `automerge:"routes" json:"routes"`
}

// MeshCrdt: Represents the mesh network as a whole
type MeshCrdt struct {
	Nodes map[string]MeshNodeCrdt `automerge:"nodes" json:"nodes"`
}
