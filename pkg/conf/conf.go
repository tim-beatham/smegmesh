// conf defines configuration file parsing for golang
package conf

import (
	"os"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"gopkg.in/yaml.v3"
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

	return &conf, nil
}
