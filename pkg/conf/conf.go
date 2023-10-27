// conf defines configuration file parsing for golang
package conf

import (
	"os"

	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"gopkg.in/yaml.v3"
)

type WgMeshConfiguration struct {
	CertificatePath      string `yaml:"certificatePath"`
	PrivateKeyPath       string `yaml:"privateKeyPath"`
	CaCertificatePath    string `yaml:"caCertificatePath"`
	SkipCertVerification bool   `yaml:"skipCertVerification"`
	GrpcPort             string `yaml:"gRPCPort"`
	// AdvertiseRoutes advertises other meshes if the node is in multiple meshes
	AdvertiseRoutes bool `yaml:"advertiseRoutes"`
	// PublicEndpoint is the IP in which this computer is publicly reachable.
	// usecase is when the node is behind NAT.
	PublicEndpoint string `yaml:"publicEndpoint"`
}

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
