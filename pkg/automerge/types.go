package crdt

// Route: Represents a CRDT of the given route
type Route struct {
	Destination string   `automerge:"destination"`
	Path        []string `automerge:"path"`
}

// MeshNodeCrdt: Represents a CRDT for a mesh nodes
type MeshNodeCrdt struct {
	HostEndpoint string            `automerge:"hostEndpoint"`
	WgEndpoint   string            `automerge:"wgEndpoint"`
	PublicKey    string            `automerge:"publicKey"`
	WgHost       string            `automerge:"wgHost"`
	Timestamp    int64             `automerge:"timestamp"`
	Routes       map[string]Route  `automerge:"routes"`
	Alias        string            `automerge:"alias"`
	Description  string            `automerge:"description"`
	Services     map[string]string `automerge:"services"`
	Type         string            `automerge:"type"`
}

// MeshCrdt: Represents the mesh network as a whole
type MeshCrdt struct {
	Nodes map[string]MeshNodeCrdt `automerge:"nodes"`
}
