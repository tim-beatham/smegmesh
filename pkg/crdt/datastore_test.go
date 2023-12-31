package crdt

import (
	"net"
	"slices"
	"testing"
	"time"

	"github.com/tim-beatham/wgmesh/pkg/conf"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	"github.com/tim-beatham/wgmesh/pkg/mesh"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TestParams struct {
	manager   mesh.MeshProvider
	publicKey *wgtypes.Key
}

func setUpTests() *TestParams {
	advertiseRoutes := false
	advertiseDefaultRoute := false
	role := conf.PEER_ROLE
	discovery := conf.DNS_IP_DISCOVERY

	factory := &TwoPhaseMapFactory{
		Config: &conf.DaemonConfiguration{
			CertificatePath:      "/somecertificatepath",
			PrivateKeyPath:       "/someprivatekeypath",
			CaCertificatePath:    "/somecacertificatepath",
			SkipCertVerification: true,
			GrpcPort:             0,
			Timeout:              20,
			Profile:              false,
			SyncTime:             2,
			HeartBeat:            10,
			ClusterSize:          32,
			InterClusterChance:   0.15,
			BranchRate:           3,
			InfectionCount:       3,
			BaseConfiguration: conf.WgConfiguration{
				IPDiscovery:           &discovery,
				AdvertiseRoutes:       &advertiseRoutes,
				AdvertiseDefaultRoute: &advertiseDefaultRoute,
				Role:                  &role,
			},
		},
	}

	key, _ := wgtypes.GeneratePrivateKey()

	mesh, _ := factory.CreateMesh(&mesh.MeshProviderFactoryParams{
		DevName:    "bob",
		MeshId:     "meshid123",
		Client:     nil,
		Conf:       &factory.Config.BaseConfiguration,
		DaemonConf: factory.Config,
		NodeID:     "bob",
	})

	publicKey := key.PublicKey()

	return &TestParams{
		manager:   mesh,
		publicKey: &publicKey,
	}
}

func getOurNode(testParams *TestParams) *MeshNode {
	return &MeshNode{
		HostEndpoint: "public-endpoint:8080",
		WgEndpoint:   "public-endpoint:21906",
		WgHost:       "3e9a:1fb3:5e50:8173:9690:f917:b1ab:d218/128",
		PublicKey:    testParams.publicKey.String(),
		Timestamp:    time.Now().Unix(),
		Description:  "A node that we are adding",
		Type:         "peer",
	}
}

func getRandomNode() *MeshNode {
	key, _ := wgtypes.GeneratePrivateKey()
	publicKey := key.PublicKey()

	return &MeshNode{
		HostEndpoint: "public-endpoint:8081",
		WgEndpoint:   "public-endpoint:21907",
		WgHost:       "3e9a:1fb3:5e50:8173:9690:f917:b1ab:d234/128",
		PublicKey:    publicKey.String(),
		Timestamp:    time.Now().Unix(),
		Description:  "A node that we are adding",
		Type:         "peer",
	}
}

func TestAddNodeAddsTheNodesToTheStore(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))

	if !testParams.manager.NodeExists(testParams.publicKey.String()) {
		t.Fatalf(`node %s should have been added to the mesh network`, testParams.publicKey.String())
	}
}

func TestAddNodeNodeAlreadyExistsReplacesTheNode(t *testing.T) {
	TestAddNodeAddsTheNodesToTheStore(t)
	TestAddNodeAddsTheNodesToTheStore(t)
}

func TestSaveThenLoad(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())

	bytes := testParams.manager.Save()

	if err := testParams.manager.Load(bytes); err != nil {
		t.Fatalf(`error caused by loading datastore: %s`, err.Error())
	}
}

func TestHasChangesReturnsTrueWhenThereAreChangesInTheMesh(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())

	if !testParams.manager.HasChanges() {
		t.Fatalf(`mesh has change but HasChanges returned false`)
	}

	testParams.manager.SetDescription(testParams.publicKey.String(), "Bob marley")

	if !testParams.manager.HasChanges() {
		t.Fatalf(`mesh has change but HasChanges returned false`)
	}

	testParams.manager.SaveChanges()
}

func TestHasChangesWhenThereAreNoChangesInTheMeshReturnsFalse(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())
	testParams.manager.AddNode(getRandomNode())

	testParams.manager.SaveChanges()

	if testParams.manager.HasChanges() {
		t.Fatalf(`mesh has no changes but HasChanges was true`)
	}

	testParams.manager.SetDescription(testParams.publicKey.String(), "Bob marley")

	testParams.manager.SaveChanges()

	if testParams.manager.HasChanges() {
		t.Fatalf(`mesh has no changes but HasChanges was true`)
	}
}

func TestUpdateTimeStampUpdatesTheTimeStampOfTheGivenNodeIfItIsTheLeader(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))

	before, _ := testParams.manager.GetNode(testParams.publicKey.String())

	time.Sleep(1 * time.Second)

	testParams.manager.UpdateTimeStamp(testParams.publicKey.String())

	after, _ := testParams.manager.GetNode(testParams.publicKey.String())

	if before.GetTimeStamp() >= after.GetTimeStamp() {
		t.Fatalf(`before should not be after after`)
	}
}

func TestUpdateTimeStampUpdatesTheTimeStampOfTheGivenNodeIfItIsNotLeader(t *testing.T) {
	testParams := setUpTests()
	testParams.manager.AddNode(getOurNode(testParams))

	newNode := getRandomNode()
	newNode.PublicKey = "aaaaaaaaaa"

	testParams.manager.AddNode(newNode)

	before, _ := testParams.manager.GetNode(testParams.publicKey.String())

	time.Sleep(1 * time.Second)

	after, _ := testParams.manager.GetNode(testParams.publicKey.String())

	if before.GetTimeStamp() != after.GetTimeStamp() {
		t.Fatalf(`before and after should be the same`)
	}
}

func TestAddRoutesAddsARouteToTheGivenMesh(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))

	_, destination, _ := net.ParseCIDR("0353:1da7:7f33:acc0:7a3f:6e55:912b:bc1f/64")

	testParams.manager.AddRoutes(testParams.publicKey.String(), &mesh.RouteStub{
		Destination: destination,
		HopCount:    0,
		Path:        make([]string, 0),
	})

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	containsDestination := lib.Contains(node.GetRoutes(), func(r mesh.Route) bool {
		return r.GetDestination().Contains(destination.IP)
	})

	if !containsDestination {
		t.Fatalf(`route has not been added to the node`)
	}
}

func TestRemoveRoutesWithdrawsRoutesFromTheMesh(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))

	_, destination, _ := net.ParseCIDR("0353:1da7:7f33:acc0:7a3f:6e55:912b:bc1f/64")
	route := &mesh.RouteStub{
		Destination: destination,
		HopCount:    0,
		Path:        make([]string, 0),
	}

	testParams.manager.AddRoutes(testParams.publicKey.String(), route)
	testParams.manager.RemoveRoutes(testParams.publicKey.String(), route)

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	containsDestination := lib.Contains(node.GetRoutes(), func(r mesh.Route) bool {
		return r.GetDestination().Contains(destination.IP)
	})

	if containsDestination {
		t.Fatalf(`route has not been removed from the node`)
	}
}

func TestGetNodeGetsTheNodeWhenItExists(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	if node == nil {
		t.Fatalf(`node not found returned nil`)
	}
}

func TestGetNodeReturnsNilWhenItDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	testParams.manager.RemoveNode(testParams.publicKey.String())

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	if node != nil {
		t.Fatalf(`node found but should be nil`)
	}
}

func TestNodeExistsReturnsFalseWhenNotExists(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))
	testParams.manager.RemoveNode(testParams.publicKey.String())

	if testParams.manager.NodeExists(testParams.publicKey.String()) {
		t.Fatalf(`nodeexists should be false`)
	}
}

func TestSetDescriptionReturnsErrorWhenNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	err := testParams.manager.SetDescription("djdjdj", "djdsjkd")

	if err == nil {
		t.Fatalf(`error should be thrown`)
	}
}

func TestSetDescriptionSetsTheDescription(t *testing.T) {
	testParams := setUpTests()
	descriptionToSet := "djdsjkd"
	testParams.manager.AddNode(getOurNode(testParams))
	err := testParams.manager.SetDescription(testParams.publicKey.String(), descriptionToSet)

	if err != nil {
		t.Fatalf(`error %s thrown`, err.Error())
	}

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	description := node.GetDescription()

	if description != descriptionToSet {
		t.Fatalf(`description was %s should be %s`, description, descriptionToSet)
	}
}

func TestAliasNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	err := testParams.manager.SetAlias("djdjdj", "djdsjkd")

	if err == nil {
		t.Fatalf(`error should be thrown`)
	}
}

func TestSetAliasSetsAlias(t *testing.T) {
	testParams := setUpTests()
	aliasToSet := "djdsjkd"
	testParams.manager.AddNode(getOurNode(testParams))
	err := testParams.manager.SetAlias(testParams.publicKey.String(), aliasToSet)

	if err != nil {
		t.Fatalf(`error %s thrown`, err.Error())
	}

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	alias := node.GetAlias()

	if alias != aliasToSet {
		t.Fatalf(`description was %s should be %s`, alias, aliasToSet)
	}
}

func TestAddServiceNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	err := testParams.manager.AddService("djdjdj", "djdsjkd", "sddsds")

	if err == nil {
		t.Fatalf(`error should be thrown`)
	}
}

func TestAddServiceNodeExists(t *testing.T) {
	testParams := setUpTests()
	service := "djdsjkd"
	serviceValue := "dsdsds"
	testParams.manager.AddNode(getOurNode(testParams))
	err := testParams.manager.AddService(testParams.publicKey.String(), service, serviceValue)

	if err != nil {
		t.Fatalf(`error %s thrown`, err.Error())
	}

	node, _ := testParams.manager.GetNode(testParams.publicKey.String())

	services := node.GetServices()

	if value, ok := services[service]; !ok || value != serviceValue {
		t.Fatalf(`service not added to the data store`)
	}
}

func TestRemoveServiceDoesNotExists(t *testing.T) {
	testParams := setUpTests()

	err := testParams.manager.RemoveService("djdjdj", "dsdssd")

	if err == nil {
		t.Fatalf(`error should be thrown`)
	}
}

func TestRemoveServiceServiceDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	testParams.manager.AddNode(getOurNode(testParams))

	if err := testParams.manager.RemoveService(testParams.publicKey.String(), "dhsdh"); err == nil {
		t.Fatalf(`error should be thrown`)
	}
}

func TestGetPeersReturnsAllPeersInTheMesh(t *testing.T) {
	testParams := setUpTests()

	peer1 := getRandomNode()
	peer2 := getRandomNode()
	client := getRandomNode()
	client.Type = "client"

	testParams.manager.AddNode(peer1)
	testParams.manager.AddNode(peer2)
	testParams.manager.AddNode(client)

	peers := testParams.manager.GetPeers()
	slices.Sort(peers)

	if len(peers) != 2 {
		t.Fatalf(`there should be two peers in the mesh`)
	}

	peer1Pub, _ := peer1.GetPublicKey()

	if !slices.Contains(peers, peer1Pub.String()) {
		t.Fatalf(`peer1 not in the list`)
	}

	peer2Pub, _ := peer2.GetPublicKey()

	if !slices.Contains(peers, peer2Pub.String()) {
		t.Fatalf(`peer2 not in the list`)
	}
}

func TestRemoveNodeReturnsErrorIfNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	err := testParams.manager.RemoveNode("dsjdssjk")

	if err == nil {
		t.Fatalf(`error should have returned`)
	}
}
