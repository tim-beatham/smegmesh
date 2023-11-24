// conf defines configuration file parsing for golang
package conf

import (
	"os"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
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
	PUBLIC_IP_DISCOVERY = "public"
	DNS_IP_DISCOVERY    = "dns"
)

type WgMeshConfiguration struct {
	// CertificatePath is the path to the certificate to use in mTLS
	CertificatePath string `yaml:"certificatePath"`
	// PrivateKeypath is the path to the clients private key in mTLS
	PrivateKeyPath string `yaml:"privateKeyPath"`
	// CaCeritifcatePath path to the certificate of the trust certificate authority
	CaCertificatePath string `yaml:"caCertificatePath"`
	// SkipCertVerification specify to skip certificate verification. Should only be used
	// in test environments
	SkipCertVerification bool `yaml:"skipCertVerification"`
	// Port to run the GrpcServer on
	GrpcPort string `yaml:"gRPCPort"`
	// IPDIscovery: how to discover your IP if not specified. Use DNS server 8.8.8.8 or
	// use public IP discovery library
	IPDiscovery IPDiscovery `yaml:"ipDiscovery"`
	// AdvertiseRoutes advertises other meshes if the node is in multiple meshes
	AdvertiseRoutes bool `yaml:"advertiseRoutes"`
	// Endpoint is the IP in which this computer is publicly reachable.
	// usecase is when the node has multiple IP addresses
	Endpoint string `yaml:"publicEndpoint"`
	// ClusterSize size of the cluster to split on
	ClusterSize int `yaml:"clusterSize"`
	// SyncRate number of times per second to perform a sync
	SyncRate float64 `yaml:"syncRate"`
	// InterClusterChance proability of inter-cluster communication in a sync round
	InterClusterChance float64 `yaml:"interClusterChance"`
	// BranchRate number of nodes to randomly communicate with
	BranchRate int `yaml:"branchRate"`
	// InfectionCount number of times we sync before we can no longer catch the udpate
	InfectionCount int `yaml:"infectionCount"`
	// KeepAliveTime number of seconds before we update node indicating that we are still alive
	KeepAliveTime int `yaml:"keepAliveTime"`
	// Timeout number of seconds before we consider the node as dead
	Timeout int `yaml:"timeout"`
	// PruneTime number of seconds before we remove nodes that are likely to be dead
	PruneTime int `yaml:"pruneTime"`
	// DeadTime: number of seconds before we consider the node as dead and stop considering it
	// when picking a random peer
	DeadTime int `yaml:"deadTime"`
	// Profile whether or not to include a http server that profiles the code
	Profile bool `yaml:"profile"`
	// StubWg whether or not to stub the WireGuard types
	StubWg bool `yaml:"stubWg"`
	// Role specifies whether or not the user is globally accessible.
	// If the user is globaly accessible they specify themselves as a client.
	Role NodeType `yaml:"role"`
	// KeepAliveWg configures the implementation so that we send keep alive packets to peers.
	// KeepAlive can only be set if role is type client
	KeepAliveWg int `yaml:"keepAliveWg"`
}

func ValidateConfiguration(c *WgMeshConfiguration) error {
	if len(c.CertificatePath) == 0 {
		return &WgMeshConfigurationError{
			msg: "A public certificate must be specified for mTLS",
		}
	}

	if len(c.PrivateKeyPath) == 0 {
		return &WgMeshConfigurationError{
			msg: "A private key must be specified for mTLS",
		}
	}

	if len(c.CaCertificatePath) == 0 {
		return &WgMeshConfigurationError{
			msg: "A ca certificate must be specified for mTLS",
		}
	}

	if len(c.GrpcPort) == 0 {
		return &WgMeshConfigurationError{
			msg: "A grpc port must be specified",
		}
	}

	if c.ClusterSize <= 0 {
		return &WgMeshConfigurationError{
			msg: "A cluster size must not be 0",
		}
	}

	if c.SyncRate <= 0 {
		return &WgMeshConfigurationError{
			msg: "SyncRate cannot be negative",
		}
	}

	if c.BranchRate <= 0 {
		return &WgMeshConfigurationError{
			msg: "Branch rate cannot be negative",
		}
	}

	if c.InfectionCount <= 0 {
		return &WgMeshConfigurationError{
			msg: "Infection count cannot be less than 1",
		}
	}

	if c.KeepAliveTime <= 0 {
		return &WgMeshConfigurationError{
			msg: "KeepAliveRate cannot be less than negative",
		}
	}

	if c.InterClusterChance <= 0 {
		return &WgMeshConfigurationError{
			msg: "Intercluster chance cannot be less than 0",
		}
	}

	if c.Timeout < 1 {
		return &WgMeshConfigurationError{
			msg: "Timeout should be greater than or equal to 1",
		}
	}

	if c.PruneTime < 1 {
		return &WgMeshConfigurationError{
			msg: "Prune time cannot be < 1",
		}
	}

	if c.DeadTime < 1 {
		return &WgMeshConfigurationError{
			msg: "Dead time cannot be < 1",
		}
	}

	if c.KeepAliveTime <= 1 {
		return &WgMeshConfigurationError{
			msg: "Prune time cannot be less than keep alive time",
		}
	}

	if c.Role == "" {
		c.Role = PEER_ROLE
	}

	if c.IPDiscovery == "" {
		c.IPDiscovery = PUBLIC_IP_DISCOVERY
	}

	return nil
}

// ParseConfiguration parses the mesh configuration
func ParseConfiguration(filePath string) (*WgMeshConfiguration, error) {
	var conf WgMeshConfiguration

	yamlBytes, err := os.ReadFile(filePath)

	if err != nil {
		logging.Log.WriteErrorf("Read file error: %s\n", err.Error())
		return nil, err
	}

	err = yaml.Unmarshal(yamlBytes, &conf)

	if err != nil {
		logging.Log.WriteErrorf("Unmarshal error: %s\n", err.Error())
		return nil, err
	}

	return &conf, ValidateConfiguration(&conf)
}
