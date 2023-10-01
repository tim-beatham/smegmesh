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
}

func ParseConfiguration(filePath string) (*WgMeshConfiguration, error) {
	var conf WgMeshConfiguration

	yamlBytes, err := os.ReadFile(filePath)

	if err != nil {
		logging.ErrorLog.Printf("Read file error: %s\n", err.Error())
		return nil, err
	}

	err = yaml.Unmarshal(yamlBytes, &conf)

	if err != nil {
		logging.ErrorLog.Printf("Unmarshal error: %s\n", err.Error())
		return nil, err
	}

	return &conf, nil
}
