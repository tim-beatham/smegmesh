package mesh

type OnChange = func([]MeshNode)

type MeshMonitor interface {
	AddUpdateCallback(cb OnChange)
	AddRemoveCallback(cb OnChange)
	Trigger() error
}

type MeshMonitorImpl struct {
	updateCbs []OnChange
	removeCbs []OnChange
	nodes     map[string]MeshNode
	manager   MeshManager
}

// Trigger causes the mesh monitor to trigger all of
// the callbacks.
func (m *MeshMonitorImpl) Trigger() error {
	changedNodes := make([]MeshNode, 0)
	removedNodes := make([]MeshNode, 0)

	nodes := make(map[string]MeshNode)

	for _, mesh := range m.manager.GetMeshes() {
		snapshot, err := mesh.GetMesh()

		if err != nil {
			return err
		}

		for _, node := range snapshot.GetNodes() {
			previous, exists := m.nodes[node.GetWgHost().String()]

			if !exists || !NodeEquals(previous, node) {
				changedNodes = append(changedNodes, node)
			}

			nodes[node.GetWgHost().String()] = node
		}
	}

	for _, previous := range m.nodes {
		_, ok := nodes[previous.GetWgHost().String()]

		if !ok {
			removedNodes = append(removedNodes, previous)
		}
	}

	if len(removedNodes) > 0 {
		for _, cb := range m.removeCbs {
			cb(removedNodes)
		}
	}

	if len(changedNodes) > 0 {
		for _, cb := range m.updateCbs {
			cb(changedNodes)
		}
	}

	return nil
}

func (m *MeshMonitorImpl) AddUpdateCallback(cb OnChange) {
	m.updateCbs = append(m.updateCbs, cb)
}

func (m *MeshMonitorImpl) AddRemoveCallback(cb OnChange) {
	m.removeCbs = append(m.removeCbs, cb)
}

func NewMeshMonitor(manager MeshManager) MeshMonitor {
	return &MeshMonitorImpl{
		updateCbs: make([]OnChange, 0),
		nodes:     make(map[string]MeshNode),
		manager:   manager,
	}
}
