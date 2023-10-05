package crdt

import "github.com/automerge/automerge-go"

// CrdtNodeManager manages nodes in the crdt mesh
type CrdtNodeManager struct {
	meshId string
	IfName string
	doc    *automerge.Doc
}

func (c *CrdtNodeManager) AddNode(crdt MeshNodeCrdt) {
	c.doc.Path("nodes").Map().Set(crdt.PublicKey, crdt)
}

// GetCrdt(): Converts the document into a struct
func (c *CrdtNodeManager) GetCrdt() (*MeshCrdt, error) {
	return automerge.As[*MeshCrdt](c.doc.Root())
}

// Load: Load an entire mesh network
func (c *CrdtNodeManager) Load(bytes []byte) error {
	doc, err := automerge.Load(bytes)

	if err != nil {
		return err
	}

	c.doc = doc
	return nil
}

// Save: Save an entire mesh network
func (c *CrdtNodeManager) Save(doc []byte) []byte {
	return c.doc.Save()
}

func (c *CrdtNodeManager) LoadChanges(changes []byte) {
	c.doc.LoadIncremental(changes)
}

func (c *CrdtNodeManager) SaveChanges() []byte {
	return c.SaveChanges()
}

// NewCrdtNodeManager: Create a new crdt node manager
func NewCrdtNodeManager(meshId, devName string) *CrdtNodeManager {
	var manager CrdtNodeManager
	manager.meshId = meshId
	manager.doc = automerge.New()
	manager.IfName = devName
	return &manager
}
