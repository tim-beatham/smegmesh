package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	secretKey     string
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
	return &JwtManager{secretKey, tokenDuration, meshes}
}

func (m *JwtManager) CreateClaims(meshId string, alias string) (*string, error) {
	node := JwtNode{
		MeshId: meshId,
		Alias:  alias,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
		},
	}

	mesh, contains := m.meshes[meshId]

	if !contains {
		return nil, errors.New("The specified mesh does not exist")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, node)
	signedString, err := token.SignedString([]byte(m.secretKey))

	if err != nil {
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
		return []byte(m.secretKey), nil
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
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]

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
