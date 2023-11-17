package mesh

type OnChange = func(string, MeshSnapshot)

type MeshMonitor interface {
	AddCallback(cb OnChange)
	Trigger(meshid string, m MeshSnapshot)
}

type MeshMonitorImpl struct {
	callbacks []OnChange
}

func (m *MeshMonitorImpl) Trigger(meshid string, snapshot MeshSnapshot) {
	for _, cb := range m.callbacks {
		cb(meshid, snapshot)
	}
}

func (m *MeshMonitorImpl) AddCallback(cb OnChange) {
	m.callbacks = append(m.callbacks, cb)
}

func NewMeshMonitor() MeshMonitor {
	return &MeshMonitorImpl{
		callbacks: make([]OnChange, 0),
	}
}
