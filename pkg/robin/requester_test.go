package robin

import (
	"testing"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/ctrlserver"
	"github.com/tim-beatham/smegmesh/pkg/ipc"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
)

func getRequester() *IpcHandler {
	return &IpcHandler{Server: ctrlserver.NewCtrlServerStub()}
}

func TestCreateMeshRepliesMeshId(t *testing.T) {
	var reply string
	requester := getRequester()

	err := requester.CreateMesh(&ipc.NewMeshArgs{
		WgArgs: ipc.WireGuardArgs{
			WgPort:   500,
			Endpoint: "abc.com:1234",
			Role:     "peer",
		},
	}, &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply) == 0 {
		t.Fatalf(`reply should have been returned`)
	}
}

func TestListMeshesNoMeshesListsEmpty(t *testing.T) {
	var reply ipc.ListMeshReply
	requester := getRequester()

	err := requester.ListMeshes("", &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply.Meshes) != 0 {
		t.Fatalf(`meshes should be empty`)
	}
}

func TestListMeshesMeshesNotEmpty(t *testing.T) {
	var reply ipc.ListMeshReply
	requester := getRequester()

	requester.Server.GetMeshManager().AddMesh(&mesh.AddMeshParams{
		MeshId:    "tim123",
		MeshBytes: make([]byte, 0),
		Conf:      &conf.WgConfiguration{},
	})

	err := requester.ListMeshes("", &reply)

	if err != nil {
		t.Error(err)
	}

	if len(reply.Meshes) != 1 {
		t.Fatalf(`only only mesh exists`)
	}

	if reply.Meshes[0] != "tim123" {
		t.Fatalf(`meshId was %s expected %s`, reply.Meshes[0], "tim123")
	}
}
