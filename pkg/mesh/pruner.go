package mesh

import (
	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
)

func pruneFunction(m MeshManager) lib.TimerFunc {
	return func() error {
		return m.Prune()
	}
}

func NewPruner(m MeshManager, conf conf.WgMeshConfiguration) *lib.Timer {
	return lib.NewTimer(pruneFunction(m), conf.PruneTime/2)
}
