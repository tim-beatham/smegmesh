package sync

// Syncer: picks random nodes from the mesh
type Syncer interface {
	Sync(meshId string) error
	SyncMeshes() error
}

type SyncerImpl struct {
}

// Sync: Sync random nodes
func (s *SyncerImpl) Sync(meshId string) error {
	return nil
}

// SyncMeshes:
func (s *SyncerImpl) SyncMeshes() error {
	return nil
}
