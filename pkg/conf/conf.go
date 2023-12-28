// conf defines configuration file parsing for golang
package conf

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type WgMeshConfigurationError struct {
	msg string
}

func (m *WgMeshConfigurationError) Error() string {
	return m.msg
}

type NodeType string

const (
	PEER_ROLE   NodeType = "peer"
	CLIENT_ROLE NodeType = "client"
)

type IPDiscovery string

const (
	PUBLIC_IP_DISCOVERY IPDiscovery = "public"
	DNS_IP_DISCOVERY    IPDiscovery = "dns"
)

// WgConfiguration contains per-mesh WireGuard configuration. Contains poitner types only so we can
// tell if the attribute is set
type WgConfiguration struct {
	// IPDIscovery: how to discover your IP if not specified. Use your outgoing IP or use a public
	// service for IPDiscoverability
	IPDiscovery *IPDiscovery `yaml:"ipDiscovery" validate:"required,eq=public|eq=dns"`
	// AdvertiseRoutes: specifies whether the node can act as a router routing packets between meshes
	AdvertiseRoutes *bool `yaml:"advertiseRoute" validate:"required"`
	// AdvertiseDefaultRoute: specifies whether or not this route should advertise a default route
	// for all nodes to route their packets to
	AdvertiseDefaultRoute *bool `yaml:"advertiseDefaults" validate:"required"`
	// Endpoint contains what value should be set as the public endpoint of this node
	Endpoint *string `yaml:"publicEndpoint"`
	// Role specifies whether or not the user is globally accessible.
	// If the user is globaly accessible they specify themselves as a client.
	Role *NodeType `yaml:"role" validate:"required,eq=client|eq=peer"`
	// KeepAliveWg configures the implementation so that we send keep alive packets to peers.
	// KeepAlive can only be set if role is type client
	KeepAliveWg *int `yaml:"keepAliveWg" validate:"omitempty,gte=0"`
	// PreUp are WireGuard commands to run before adding the WG interface
	PreUp []string `yaml:"preUp"`
	// PostUp are WireGuard commands to run after adding the WG interface
	PostUp []string `yaml:"postUp"`
	// PreDown are WireGuard commands to run prior to removing the WG interface
	PreDown []string `yaml:"preDown"`
	// PostDown are WireGuard command to run after removing the WG interface
	PostDown []string `yaml:"postDown"`
}

type DaemonConfiguration struct {
	// CertificatePath is the path to the certificate to use in mTLS
	CertificatePath string `yaml:"certificatePath" validate:"required"`
	// PrivateKeypath is the path to the clients private key in mTLS
	PrivateKeyPath string `yaml:"privateKeyPath" validate:"required"`
	// CaCeritifcatePath path to the certificate of the trust certificate authority
	CaCertificatePath string `yaml:"caCertificatePath" validate:"required"`
	// SkipCertVerification specify to skip certificate verification. Should only be used
	// in test environments
	SkipCertVerification bool `yaml:"skipCertVerification"`
	// Port to run the GrpcServer on
	GrpcPort int `yaml:"gRPCPort" validate:"required"`
	// Timeout number of seconds without response that a node is considered unreachable by gRPC
	Timeout int `yaml:"timeout" validate:"required,gte=1"`
	// Profile whether or not to include a http server that profiles the code
	Profile bool `yaml:"profile"`
	// StubWg whether or not to stub the WireGuard types
	StubWg bool `yaml:"stubWg"`
	// SyncRate specifies how long the minimum time should be between synchronisation
	SyncRate int `yaml:"syncRate" validate:"required,gte=1"`
	// KeepAliveTime: number of seconds before the leader of the mesh sends an update to
	// send to every member in the mesh
	KeepAliveTime int `yaml:"keepAliveTime" validate:"required,gte=1"`
	// ClusterSize specifies how many neighbours you should synchronise with per round
	ClusterSize int `yaml:"clusterSize" validate:"gte=1"`
	// InterClusterChance specifies the probabilityof inter-cluster communication in a sync round
	InterClusterChance float64 `yaml:"interClusterChance" validate:"gt=0"`
	// BranchRate specifies the number of nodes to synchronise with when a node has
	// new changes to send to the mesh
	BranchRate int `yaml:"branchRate" validate:"required,gte=1"`
	// InfectionCount: number of time to sync before an update can no longer be 'caught'
	InfectionCount int `yaml:"infectionCount" validate:"required,gte=1"`
	// BaseConfiguration base WireGuard configuration to use, this is used when none is provided
	BaseConfiguration WgConfiguration `yaml:"baseConfiguration" validate:"required"`
}

// ValdiateMeshConfiguration: validates the mesh configuration
func ValidateMeshConfiguration(conf *WgConfiguration) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(conf)

	if conf.PostDown == nil {
		conf.PostDown = make([]string, 0)
	}

	if conf.PostUp == nil {
		conf.PostUp = make([]string, 0)
	}

	if conf.PreDown == nil {
		conf.PreDown = make([]string, 0)
	}

	if conf.PreUp == nil {
		conf.PreUp = make([]string, 0)
	}

	return err
}

// ValidateDaemonConfiguration: validates the dameon configuration that is used.
func ValidateDaemonConfiguration(c *DaemonConfiguration) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(c)
	return err
}

// ParseDaemonConfiguration parses the mesh configuration and validates the configuration
func ParseDaemonConfiguration(filePath string) (*DaemonConfiguration, error) {
	var conf DaemonConfiguration

	yamlBytes, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlBytes, &conf)

	if err != nil {
		return nil, err
	}

	if conf.BaseConfiguration.KeepAliveWg == nil {
		var keepAlive int = 0
		conf.BaseConfiguration.KeepAliveWg = &keepAlive
	}

	return &conf, ValidateDaemonConfiguration(&conf)
}

// MergemeshConfiguration: merges the configuration in precedence where the last
// element in the list takes the most and the first takes the least
func MergeMeshConfiguration(cfgs ...WgConfiguration) (WgConfiguration, error) {
	var result WgConfiguration

	for _, cfg := range cfgs {
		if cfg.AdvertiseDefaultRoute != nil {
			result.AdvertiseDefaultRoute = cfg.AdvertiseDefaultRoute
		}

		if cfg.AdvertiseRoutes != nil {
			result.AdvertiseRoutes = cfg.AdvertiseRoutes
		}

		if cfg.Endpoint != nil {
			result.Endpoint = cfg.Endpoint
		}

		if cfg.IPDiscovery != nil {
			result.IPDiscovery = cfg.IPDiscovery
		}

		if cfg.KeepAliveWg != nil {
			result.KeepAliveWg = cfg.KeepAliveWg
		}

		if cfg.PostDown != nil {
			result.PostDown = cfg.PostDown
		}

		if cfg.PostUp != nil {
			result.PostUp = cfg.PostUp
		}

		if cfg.PreDown != nil {
			result.PreDown = cfg.PreDown
		}

		if cfg.PreUp != nil {
			result.PreUp = cfg.PreUp
		}

		if cfg.Role != nil {
			result.Role = cfg.Role
		}
	}

	return result, ValidateMeshConfiguration(&result)
}
