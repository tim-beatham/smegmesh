package conf

import (
	"testing"
)

func getExampleConfiguration() *DaemonConfiguration {
	discovery := PUBLIC_IP_DISCOVERY
	advertiseRoutes := false
	advertiseDefaultRoute := false
	endpoint := "abc.com:123"
	nodeType := CLIENT_ROLE
	keepAliveWg := 0

	return &DaemonConfiguration{
		CertificatePath:      "../../../cert/cert.pem",
		PrivateKeyPath:       "../../../cert/priv.pem",
		CaCertificatePath:    "../../../cert/cacert.pem",
		SkipCertVerification: true,
		GrpcPort:             25,
		Timeout:              5,
		Profile:              false,
		StubWg:               false,
		SyncTime:             2,
		HeartBeat:            2,
		ClusterSize:          64,
		InterClusterChance:   0.15,
		BranchRate:           3,
		PullTime:             0,
		InfectionCount:       2,
		BaseConfiguration: WgConfiguration{
			IPDiscovery:           &discovery,
			AdvertiseRoutes:       &advertiseRoutes,
			AdvertiseDefaultRoute: &advertiseDefaultRoute,
			Endpoint:              &endpoint,
			Role:                  &nodeType,
			KeepAliveWg:           &keepAliveWg,
		},
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

func TestIPDiscoveryNotSet(t *testing.T) {
	conf := getExampleConfiguration()
	ipDiscovery := IPDiscovery("djdsjdskd")
	conf.BaseConfiguration.IPDiscovery = &ipDiscovery

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestAdvertiseRoutesNotSet(t *testing.T) {
	conf := getExampleConfiguration()
	conf.BaseConfiguration.AdvertiseRoutes = nil

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestAdvertiseDefaultRouteNotSet(t *testing.T) {
	conf := getExampleConfiguration()
	conf.BaseConfiguration.AdvertiseDefaultRoute = nil

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestKeepAliveWgNegative(t *testing.T) {
	conf := getExampleConfiguration()
	keepAliveWg := -1
	conf.BaseConfiguration.KeepAliveWg = &keepAliveWg

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestRoleTypeNotValid(t *testing.T) {
	conf := getExampleConfiguration()
	role := NodeType("bruhhh")
	conf.BaseConfiguration.Role = &role

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestRoleTypeNotSpecified(t *testing.T) {
	conf := getExampleConfiguration()
	conf.BaseConfiguration.Role = nil

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`invalid role type`)
	}
}

func TestBranchRateZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.BranchRate = 0

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestsyncTimeZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.SyncTime = 0

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestKeepAliveTimeZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.HeartBeat = 0
	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestClusterSizeZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.ClusterSize = 0
	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestInterClusterChanceZero(t *testing.T) {
	conf := getExampleConfiguration()
	conf.InterClusterChance = 0

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestInfectionCountOne(t *testing.T) {
	conf := getExampleConfiguration()
	conf.InfectionCount = 0

	err := ValidateDaemonConfiguration(conf)

	if err == nil {
		t.Fatal(`error should be thrown`)
	}
}

func TestPullTimeNegative(t *testing.T) {
	conf := getExampleConfiguration()
	conf.PullTime = -1

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
