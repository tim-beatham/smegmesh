package crdt

type MeshNodeCrdt struct {
	HostEndpoint string `automerge:"hostEndpoint"`
	WgEndpoint   string `automerge:"wgEndpoint"`
	PublicKey    string `automerge:"publicKey"`
	WgHost       string `automerge:"wgHost"`
}

type MeshCrdt struct {
	Nodes map[string]MeshNodeCrdt `automerge:"nodes"`
}
