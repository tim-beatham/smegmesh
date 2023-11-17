package crdt

// MeshNodeCrdt: Represents a CRDT for a mesh nodes
type MeshNodeCrdt struct {
	HostEndpoint string                 `automerge:"hostEndpoint"`
	WgEndpoint   string                 `automerge:"wgEndpoint"`
	PublicKey    string                 `automerge:"publicKey"`
	WgHost       string                 `automerge:"wgHost"`
	Timestamp    int64                  `automerge:"timestamp"`
	Routes       map[string]interface{} `automerge:"routes"`
	Alias        string                 `automerge:"alias"`
	Description  string                 `automerge:"description"`
	Services     map[string]string      `automerge:"services"`
}

// MeshCrdt: Represents the mesh network as a whole
type MeshCrdt struct {
	Nodes map[string]MeshNodeCrdt `automerge:"nodes"`
}
