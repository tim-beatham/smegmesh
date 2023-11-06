package conf

import "testing"

func getExampleConfiguration() *WgMeshConfiguration {
	return &WgMeshConfiguration{
		CertificatePath:      "./cert/cert.pem",
		PrivateKeyPath:       "./cert/key.pem",
		CaCertificatePath:    "./cert/ca.pems",
		SkipCertVerification: true,
		GrpcPort:             "8080",
		AdvertiseRoutes:      true,
		Endpoint:             "localhost",
		ClusterSize:          1,
		SyncRate:             1,
		InterClusterChance:   0.1,
		BranchRate:           2,
		KeepAliveTime:        4,
		InfectionCount:       1,
		Timeout:              2,
		PruneTime:            20,
	}
}

func TestConfigurationCertificatePathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.CertificatePath = ""

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationPrivateKeyPathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.PrivateKeyPath = ""

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationCaCertificatePathEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.CaCertificatePath = ""

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestConfigurationGrpcPortEmpty(t *testing.T) {
	conf := getExampleConfiguration()
	conf.GrpcPort = ""

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestClusterSizeZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.ClusterSize = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func SyncRateZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.SyncRate = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func BranchRateZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.BranchRate = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func InfectionCountZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.InfectionCount = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func KeepAliveRateZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.KeepAliveTime = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestValidCOnfiguration(t *testing.T) {
	conf := getExampleConfiguration()

	err := ValidateConfiguration(conf)

	if err != nil {
		t.Error(err)
	}
}

func TestTimeout(t *testing.T) {
	conf := getExampleConfiguration()
	conf.Timeout = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestPruneTimeZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.PruneTime = 0

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatalf(`Error should be thrown`)
	}
}

func TestPruneTimeLessThanKeepAliveTime(t *testing.T) {
	conf := getExampleConfiguration()
	conf.PruneTime = 1

	err := ValidateConfiguration(conf)

	if err == nil {
		t.Fatalf(`Error should be thrown`)
	}
}
