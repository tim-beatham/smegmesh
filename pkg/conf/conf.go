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
	SkipCertVerification bool   `yaml:"skipCertVerification"`
	IfName               string `yaml:"ifName"`
	WgPort               int    `yaml:"wgPort"`
	GrpcPort             string `yaml:"gRPCPort"`
	Secret               string `yaml:"secret"`
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
