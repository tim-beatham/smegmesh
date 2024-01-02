package automerge

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tim-beatham/smegmesh/pkg/conf"
	"github.com/tim-beatham/smegmesh/pkg/lib"
	"github.com/tim-beatham/smegmesh/pkg/mesh"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TestParams struct {
	manager *CrdtMeshManager
}

func setUpTests() *TestParams {
	manager, _ := NewCrdtNodeManager(&NewCrdtNodeMangerParams{
		MeshId:  "timsmesh123",
		DevName: "wg0",
		Port:    5000,
		Client:  nil,
		Conf:    &conf.WgConfiguration{},
	})

	return &TestParams{
		manager: manager,
	}
}

func getTestNode() mesh.MeshNode {
	pubKey, _ := wgtypes.GeneratePrivateKey()

	return &MeshNodeCrdt{
		HostEndpoint: "public-endpoint:8080",
		WgEndpoint:   "public-endpoint:21906",
		WgHost:       "3e9a:1fb3:5e50:8173:9690:f917:b1ab:d218/128",
		PublicKey:    pubKey.String(),
		Timestamp:    time.Now().Unix(),
		Description:  "A node that we are adding",
	}
}

func getTestNode2() mesh.MeshNode {
	pubKey, _ := wgtypes.GeneratePrivateKey()

	return &MeshNodeCrdt{
		HostEndpoint: "public-endpoint:8081",
		WgEndpoint:   "public-endpoint:21907",
		WgHost:       "3e9a:1fb3:5e50:8173:9690:f917:b1ab:d219/128",
		PublicKey:    pubKey.String(),
		Timestamp:    time.Now().Unix(),
		Description:  "A node that we are adding",
	}
}

func TestAddNodeNodeExists(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()
	testParams.manager.AddNode(node)

	pubKey, _ := node.GetPublicKey()
	node, err := testParams.manager.GetNode(pubKey.String())

	if err != nil {
		t.Error(err)
	}

	if node == nil {
		t.Fatalf(`node not added to the mesh when it should be`)
	}
}

func TestAddNodeAddRoute(t *testing.T) {
	testParams := setUpTests()
	testNode := getTestNode()
	pubKey, _ := testNode.GetPublicKey()

	_, destination, _ := net.ParseCIDR("fd:1c64:1d00::/48")

	testParams.manager.AddNode(testNode)
	testParams.manager.AddRoutes(pubKey.String(), &mesh.RouteStub{
		Destination: destination,
		HopCount:    0,
		Path:        make([]string, 0),
	})
	updatedNode, err := testParams.manager.GetNode(pubKey.String())

	if err != nil {
		t.Error(err)
	}

	if updatedNode == nil {
		t.Fatalf(`node does not exist in the mesh`)
	}

	routes := updatedNode.GetRoutes()

	if len(routes) != 1 {
		t.Fatal(`Route length mismatch`)
	}
}

func TestGetMeshIdReturnsTheMeshId(t *testing.T) {
	testParams := setUpTests()

	if len(testParams.manager.GetMeshId()) == 0 {
		t.Fatal(`Meshid is less than 0`)
	}
}

// Add 2 nodes to the mesh and then get the mesh.s
// It should return the 2 nodes that have been added to the mesh
func TestAdd2NodesGetMesh(t *testing.T) {
	testParams := setUpTests()
	node1 := getTestNode()
	node2 := getTestNode2()

	testParams.manager.AddNode(node1)
	testParams.manager.AddNode(node2)

	mesh, err := testParams.manager.GetMesh()

	if err != nil {
		t.Error(err)
	}

	nodes := mesh.GetNodes()

	if len(nodes) != 2 {
		t.Fatalf(`Mismatch in node slice`)
	}

	for _, node := range nodes {
		if node.GetHostEndpoint() != node1.GetHostEndpoint() && node.GetHostEndpoint() != node2.GetHostEndpoint() {
			t.Fatalf(`Node should not exist`)
		}
	}
}

func TestSaveMeshReturnsMeshBytes(t *testing.T) {
	testParams := setUpTests()
	node1 := getTestNode()

	testParams.manager.AddNode(node1)

	bytes := testParams.manager.Save()

	if len(bytes) <= 0 {
		t.Fatalf(`bytes in the mesh is less than 0`)
	}
}

func TestSaveMeshThenLoad(t *testing.T) {
	testParams := setUpTests()
	testParams2 := setUpTests()

	node1 := getTestNode()
	testParams.manager.AddNode(node1)

	bytes := testParams.manager.Save()

	err := testParams2.manager.Load(bytes)

	if err != nil {
		t.Error(err)
	}

	if len(bytes) <= 0 {
		t.Fatalf(`bytes in the mesh is less than 0`)
	}

	mesh2, err := testParams2.manager.GetMesh()

	if err != nil {
		t.Error(err)
	}

	nodes := mesh2.GetNodes()

	if lib.MapValues(nodes)[0].GetHostEndpoint() != node1.GetHostEndpoint() {
		t.Fatalf(`Node should be in the list of nodes`)
	}
}

func TestLengthNoNodes(t *testing.T) {
	testParams := setUpTests()

	if testParams.manager.Length() != 0 {
		t.Fatalf(`Number of nodes should be 0`)
	}
}

func TestLength1Node(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()
	testParams.manager.AddNode(node)

	if testParams.manager.Length() != 1 {
		t.Fatalf(`Number of nodes should be 1`)
	}
}

func TestLengthMultipleNodes(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()
	node1 := getTestNode2()

	testParams.manager.AddNode(node)
	testParams.manager.AddNode(node1)

	if testParams.manager.Length() != 2 {
		t.Fatalf(`Number of nodes should be 2`)
	}
}

func TestHasChangesNoChanges(t *testing.T) {
	testParams := setUpTests()

	if testParams.manager.HasChanges() {
		t.Fatalf(`Should not have changes just created document`)
	}
}

func TestHasChangesChanges(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()

	testParams.manager.AddNode(node)

	if !testParams.manager.HasChanges() {
		t.Fatalf(`Should have changes just added node`)
	}
}

func TestHasChangesSavedChanges(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()

	testParams.manager.AddNode(node)

	testParams.manager.SaveChanges()

	if testParams.manager.HasChanges() {
		t.Fatalf(`Should not have changes just saved document`)
	}
}

func TestUpdateTimeStampNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()
	err := testParams.manager.UpdateTimeStamp("AAAAAA")

	if err == nil {
		t.Fatalf(`Error should have returned`)
	}
}

func TestUpdateTimeStampNodeExists(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()

	testParams.manager.AddNode(node)
	pubKey, _ := node.GetPublicKey()

	err := testParams.manager.UpdateTimeStamp(pubKey.String())

	if err != nil {
		t.Error(err)
	}
}

func TestSetDescriptionNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()
	err := testParams.manager.SetDescription("AAAAA", "Bob 123")

	if err == nil {
		t.Fatalf(`Error should have returned`)
	}
}

func TestSetDescriptionNodeExists(t *testing.T) {
	testParams := setUpTests()
	node := getTestNode()
	err := testParams.manager.SetDescription(node.GetHostEndpoint(), "Bob 123")

	if err == nil {
		t.Fatalf(`Error should have returned`)
	}
}

func TestAddRoutesNodeDoesNotExist(t *testing.T) {
	testParams := setUpTests()

	_, destination, _ := net.ParseCIDR("fd:1c64:1d00::/48")

	err := testParams.manager.AddRoutes("AAAAA", &mesh.RouteStub{
		Destination: destination,
		HopCount:    0,
		Path:        make([]string, 0),
	})

	if err == nil {
		t.Error(err)
	}
}

func TestCompareComparesByPublicKey(t *testing.T) {
	node := getTestNode().(*MeshNodeCrdt)
	node2 := getTestNode2().(*MeshNodeCrdt)

	pubKey1, _ := node.GetPublicKey()
	pubKey2, _ := node2.GetPublicKey()

	if node.Compare(node2) != strings.Compare(pubKey1.String(), pubKey2.String()) {
		t.Fatalf(`compare failed`)
	}
}

func TestGetHostEndpoint(t *testing.T) {
	node := getTestNode()

	if (node.(*MeshNodeCrdt)).HostEndpoint != node.GetHostEndpoint() {
		t.Fatalf(`get hostendpoint should get the host endpoint`)
	}
}

func TestGetPublicKey(t *testing.T) {
	key1, _ := wgtypes.GenerateKey()

	node := getTestNode()
	node.(*MeshNodeCrdt).PublicKey = key1.String()

	pubKey, err := node.GetPublicKey()

	if err != nil {
		t.Error(err)
	}

	if pubKey.String() != key1.String() {
		t.Fatalf(`Expected %s got %s`, key1.String(), pubKey.String())
	}
}

func TestGetWgEndpoint(t *testing.T) {
	node := getTestNode()

	if node.(*MeshNodeCrdt).WgEndpoint != node.GetWgEndpoint() {
		t.Fatal(`Did not return the correct wgEndpoint`)
	}
}

func TestGetWgHost(t *testing.T) {
	node := getTestNode()

	ip := node.GetWgHost()

	if node.(*MeshNodeCrdt).WgHost != ip.String() {
		t.Fatal(`Did not parse WgHost correctly`)
	}
}

func TestGetTimeStamp(t *testing.T) {
	node := getTestNode()

	if node.(*MeshNodeCrdt).Timestamp != node.GetTimeStamp() {
		t.Fatal(`Did not return return the correct timestamp`)
	}
}

func TestGetIdentifierDoesNotContainPrefix(t *testing.T) {
	node := getTestNode()

	if strings.Contains(node.GetIdentifier(), "/128") {
		t.Fatal(`Identifier should not contain prefix`)
	}
}
