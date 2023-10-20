package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JwtMesh contains all the sessions with the mesh network
type JwtMesh struct {
	meshId string
	// nodes contains a set of nodes with the string being the jwt token
	nodes map[string]interface{}
}

// JwtManager manages jwt tokens indicating a session
// between this host and another within a specific mesh
type JwtManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	// meshes contains all the meshes that we have sessions with
	meshes map[string]*JwtMesh
}

// JwtNode represents a jwt node in the mesh network
type JwtNode struct {
	MeshId string `json:"meshId"`
	Alias  string `json:"alias"`
	jwt.RegisteredClaims
}

func NewJwtManager(secretKey string, tokenDuration time.Duration) *JwtManager {
	meshes := make(map[string]*JwtMesh)
	return &JwtManager{[]byte(secretKey), tokenDuration, meshes}
}

func (m *JwtManager) CreateClaims(meshId string, alias string) (*string, error) {
	logging.InfoLog.Println("MeshID: " + meshId)
	logging.InfoLog.Println("Token Duration: " + strconv.Itoa(int(m.tokenDuration)))
	node := JwtNode{
		MeshId: meshId,
		Alias:  alias,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
		},
	}

	mesh, contains := m.meshes[meshId]

	if !contains {
		mesh = new(JwtMesh)
		mesh.meshId = meshId
		mesh.nodes = make(map[string]interface{})
		mesh.nodes[meshId] = mesh
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, node)
	signedString, err := token.SignedString(m.secretKey)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	_, exists := mesh.nodes[signedString]

	if exists {
		return nil, errors.New("Node already exists")
	}

	mesh.nodes[signedString] = struct{}{}
	return &signedString, nil
}

func (m *JwtManager) Verify(accessToken string) (*JwtNode, bool) {
	token, err := jwt.ParseWithClaims(accessToken, &JwtNode{}, func(t *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})

	if err != nil {
		return nil, false
	}

	if !token.Valid {
		return nil, token.Valid
	}

	claims, ok := token.Claims.(*JwtNode)
	return claims, ok
}

func (m *JwtManager) GetAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if strings.Contains(info.FullMethod, "") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]

		for _, w := range values {
			logging.InfoLog.Printf(w)
		}

		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		acessToken := values[0]

		_, valid := m.Verify(acessToken)

		if !valid {
			return nil, status.Errorf(codes.Unauthenticated, "Invalid access token: %s", acessToken)
		}

		return handler(ctx, req)
	}
}
