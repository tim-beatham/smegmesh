package mesh

import (
	"testing"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/wg"
)

func getMeshConfiguration() *conf.WgMeshConfiguration {
	return &conf.WgMeshConfiguration{
		GrpcPort:           "8080",
		Endpoint:           "abc.com",
		ClusterSize:        64,
		SyncRate:           4,
		BranchRate:         3,
		InterClusterChance: 0.15,
		InfectionCount:     2,
		KeepAliveTime:      60,
	}
}

func getMeshManager() MeshManager {
	manager := NewMeshManager(&NewMeshManagerParams{
		Conf:                 *getMeshConfiguration(),
		Client:               nil,
		MeshProvider:         &StubMeshProviderFactory{},
		NodeFactory:          &StubNodeFactory{Config: getMeshConfiguration()},
		IdGenerator:          &lib.UUIDGenerator{},
		IPAllocator:          &ip.ULABuilder{},
		InterfaceManipulator: &wg.WgInterfaceManipulatorStub{},
		ConfigApplyer:        &MeshConfigApplyerStub{},
		RouteManager:         &RouteManagerStub{},
	})

	return manager
}

func TestCreateMeshCreatesANewMeshProvider(t *testing.T) {
	manager := getMeshManager()

	meshId, err := manager.CreateMesh("wg0", 5000)

	if err != nil {
		t.Error(err)
	}

	if len(meshId) == 0 {
		t.Fatal(`meshId should not be empty`)
	}

	_, exists := manager.GetMeshes()[meshId]

	if !exists {
		t.Fatal(`mesh was not created when it should be`)
	}
}

func TestAddMeshAddsAMesh(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	manager.AddMesh(&AddMeshParams{
		MeshId:    meshId,
		DevName:   "wg0",
		WgPort:    6000,
		MeshBytes: make([]byte, 0),
	})

	mesh := manager.GetMesh(meshId)

	if mesh == nil || mesh.GetMeshId() != meshId {
		t.Fatalf(`mesh has not been added to the list of meshes`)
	}
}

func TestAddMeshMeshAlreadyExistsReplacesIt(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	for i := 0; i < 2; i++ {
		err := manager.AddMesh(&AddMeshParams{
			MeshId:    meshId,
			DevName:   "wg0",
			WgPort:    6000,
			MeshBytes: make([]byte, 0),
		})

		if err != nil {
			t.Error(err)
		}
	}

	mesh := manager.GetMesh(meshId)

	if mesh == nil || mesh.GetMeshId() != meshId {
		t.Fatalf(`mesh has not been added to the list of meshes`)
	}
}

func TestAddSelfAddsSelfToTheMesh(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	err := manager.AddMesh(&AddMeshParams{
		MeshId:    meshId,
		DevName:   "wg0",
		WgPort:    6000,
		MeshBytes: make([]byte, 0),
	})

	if err != nil {
		t.Error(err)
	}

	err = manager.AddSelf(&AddSelfParams{
		MeshId:   meshId,
		WgPort:   5000,
		Endpoint: "abc.com",
	})

	if err != nil {
		t.Error(err)
	}

	mesh, err := manager.GetMesh(meshId).GetMesh()

	if err != nil {
		t.Error(err)
	}

	_, ok := mesh.GetNodes()["abc.com"]

	if !ok {
		t.Fatalf(`node has not been added`)
	}
}

func TestAddSelfToMeshAlreadyInMesh(t *testing.T) {
	TestAddSelfAddsSelfToTheMesh(t)
	TestAddSelfAddsSelfToTheMesh(t)
}

func TestAddSelfToMeshMeshDoesNotExist(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	err := manager.AddSelf(&AddSelfParams{
		MeshId:   meshId,
		WgPort:   5000,
		Endpoint: "abc.com",
	})

	if err == nil {
		t.Fatalf(`Expected error to be thrown`)
	}
}

func TestLeaveMeshMeshDoesNotExist(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	err := manager.LeaveMesh(meshId)

	if err == nil {
		t.Fatalf(`Expected error to be thrown`)
	}
}

func TestLeaveMeshDeletesMesh(t *testing.T) {
	manager := getMeshManager()
	meshId := "meshid123"

	err := manager.AddMesh(&AddMeshParams{
		MeshId:    meshId,
		DevName:   "wg0",
		WgPort:    6000,
		MeshBytes: make([]byte, 0),
	})

	if err != nil {
		t.Error(err)
	}

	err = manager.LeaveMesh(meshId)

	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	_, exists := manager.GetMeshes()[meshId]

	if exists {
		t.Fatalf(`expected mesh to have been deleted`)
	}
}

func TestSetDescription(t *testing.T) {
	manager := getMeshManager()
	description := "wooooo"

	meshId1, _ := manager.CreateMesh("wg0", 5000)
	meshId2, _ := manager.CreateMesh("wg0", 5001)

	manager.AddSelf(&AddSelfParams{
		MeshId:   meshId1,
		WgPort:   5000,
		Endpoint: "abc.com:8080",
	})
	manager.AddSelf(&AddSelfParams{
		MeshId:   meshId2,
		WgPort:   5000,
		Endpoint: "abc.com:8080",
	})

	err := manager.SetDescription(description)

	if err != nil {
		t.Fatalf(`failed to set the descriptions`)
	}
}

func TestUpdateTimeStampUpdatesAllMeshes(t *testing.T) {
	manager := getMeshManager()

	meshId1, _ := manager.CreateMesh("wg0", 5000)
	meshId2, _ := manager.CreateMesh("wg0", 5001)

	manager.AddSelf(&AddSelfParams{
		MeshId:   meshId1,
		WgPort:   5000,
		Endpoint: "abc.com:8080",
	})
	manager.AddSelf(&AddSelfParams{
		MeshId:   meshId2,
		WgPort:   5000,
		Endpoint: "abc.com:8080",
	})

	err := manager.UpdateTimeStamp()

	if err != nil {
		t.Fatalf(`failed to update the timestamp`)
	}
}
