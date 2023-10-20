package crdt

import "github.com/automerge/automerge-go"

type MeshNodeCrdt struct {
	HostEndpoint string             `automerge:"hostEndpoint"`
	WgEndpoint   string             `automerge:"wgEndpoint"`
	PublicKey    string             `automerge:"publicKey"`
	WgHost       string             `automerge:"wgHost"`
	FailedCount  *automerge.Counter `automerge:"failedCount"`
	FailedInt    int                `automerge:"-"`
}

type MeshCrdt struct {
	Nodes map[string]MeshNodeCrdt `automerge:"nodes"`
}
