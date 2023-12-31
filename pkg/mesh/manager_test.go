package mesh

import (
	"testing"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/ip"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/wg"
)

func getMeshConfiguration() *conf.DaemonConfiguration {
	advertiseRoutes := true
	advertiseDefaultRoute := true
	ipDiscovery := conf.PUBLIC_IP_DISCOVERY
	role := conf.PEER_ROLE

	return &conf.DaemonConfiguration{
		GrpcPort:             8080,
		CertificatePath:      "./somecertificatepath",
		PrivateKeyPath:       "./someprivatekeypath",
		CaCertificatePath:    "./somecacertificatepath",
		SkipCertVerification: true,
		Timeout:              5,
		Profile:              false,
		StubWg:               true,
		SyncTime:             2,
		HeartBeat:            60,
		ClusterSize:          64,
		InterClusterChance:   0.15,
		BranchRate:           3,
		InfectionCount:       3,
		BaseConfiguration: conf.WgConfiguration{
			IPDiscovery:           &ipDiscovery,
			AdvertiseRoutes:       &advertiseRoutes,
			AdvertiseDefaultRoute: &advertiseDefaultRoute,
			Role:                  &role,
		},
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

	meshId, err := manager.CreateMesh(&CreateMeshParams{
		Port: 0,
		Conf: &conf.WgConfiguration{},
	})

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

	_, ok := mesh.GetNodes()[manager.GetPublicKey().String()]

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

func TestSetAlias(t *testing.T) {
	manager := getMeshManager()
	alias := "Firpo"

	meshId, _ := manager.CreateMesh(&CreateMeshParams{
		Port: 5000,
		Conf: &conf.WgConfiguration{},
	})

	manager.AddSelf(&AddSelfParams{
		MeshId:   meshId,
		WgPort:   5000,
		Endpoint: "abc.com:8080",
	})

	err := manager.SetAlias(alias)

	if err != nil {
		t.Fatalf(`failed to set the alias`)
	}

	self, err := manager.GetSelf(meshId)

	if err != nil {
		t.Fatalf(`failed to set the alias err: %s`, err.Error())
	}

	if alias != self.GetAlias() {
		t.Fatalf(`alias should be %s was %s`, alias, self.GetAlias())
	}
}

func TestSetDescription(t *testing.T) {
	manager := getMeshManager()
	description := "wooooo"

	meshId1, _ := manager.CreateMesh(&CreateMeshParams{
		Port: 5000,
		Conf: &conf.WgConfiguration{},
	})

	meshId2, _ := manager.CreateMesh(&CreateMeshParams{
		Port: 5001,
		Conf: &conf.WgConfiguration{},
	})

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

	self1, err := manager.GetSelf(meshId1)

	if err != nil {
		t.Fatalf(`failed to set the description`)
	}

	if description != self1.GetDescription() {
		t.Fatalf(`description should be %s was %s`, description, self1.GetDescription())
	}

	self2, err := manager.GetSelf(meshId2)

	if err != nil {
		t.Fatalf(`failed to set the description`)
	}

	if description != self2.GetDescription() {
		t.Fatalf(`description should be %s was %s`, description, self2.GetDescription())
	}
}

func TestUpdateTimeStampUpdatesAllMeshes(t *testing.T) {
	manager := getMeshManager()

	meshId1, _ := manager.CreateMesh(&CreateMeshParams{
		Port: 5000,
		Conf: &conf.WgConfiguration{},
	})

	meshId2, _ := manager.CreateMesh(&CreateMeshParams{
		Port: 5001,
		Conf: &conf.WgConfiguration{},
	})

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
