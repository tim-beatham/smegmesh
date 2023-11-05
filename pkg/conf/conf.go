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
	// AdvertiseRoutes advertises other meshes if the node is in multiple meshes
	AdvertiseRoutes bool `yaml:"advertiseRoutes"`
	// Endpoint is the IP in which this computer is publicly reachable.
	// usecase is when the node has multiple IP addresses
	Endpoint           string  `yaml:"publicEndpoint"`
	ClusterSize        int     `yaml:"clusterSize"`
	SyncRate           float64 `yaml:"syncRate"`
	InterClusterChance float64 `yaml:"interClusterChance"`
	BranchRate         int     `yaml:"branchRate"`
	InfectionCount     int     `yaml:"infectionCount"`
	KeepAliveRate      int     `yaml:"keepAliveRate"`
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

	if c.KeepAliveRate <= 0 {
		return &WgMeshConfigurationError{
			msg: "KeepAliveRate cannot be less than negative",
		}
	}

	if c.InterClusterChance <= 0 {
		return &WgMeshConfigurationError{
			msg: "Intercluster chance cannot be less than 0",
		}
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
