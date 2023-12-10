package conf

import "testing"

func getExampleConfiguration() *DaemonConfiguration {
	return &DaemonConfiguration{
		CertificatePath:      "./cert/cert.pem",
		PrivateKeyPath:       "./cert/key.pem",
		CaCertificatePath:    "./cert/ca.pems",
		SkipCertVerification: true,
	}
}

func TestConfigurationCertificatePathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.CertificatePath = ""

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationPrivateKeyPathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.PrivateKeyPath = ""

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationCaCertificatePathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.CaCertificatePath = ""

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationGrpcPortEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.GrpcPort = 0

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestValidConfiguration(t *testing.T) {
	conf := getExampleConfiguration()

	err := ValidateDaemonConfiguration(conf)

	if err != nil {
		t.Error(err)
	}
}
