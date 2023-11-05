package conn

import (
	"crypto/tls"
	"errors"
	"log"
	"testing"
)

func getConnectionManagerParams() *NewConnectionManagerParams {
	return &NewConnectionManagerParams{
		CertificatePath:      "./test/cert.pem",
		PrivateKey:           "./test/priv.pem",
		CaCert:               "./test/cacert.pem",
		SkipCertVerification: false,
		ConnFactory:          MockFactory,
	}
}

func TestNewConnectionManagerCertificatePathDoesNotExist(t *testing.T) {
	params := getConnectionManagerParams()
	params.CertificatePath = "./cert/sdfjdskjdsjkd.pem"

	_, err := NewConnectionManager(params)

	if err == nil {
		t.Fatalf(`Expected error as certificate does not exist`)
	}
}

func TestNewConnectionManagerPrivateKeyDoesNotExist(t *testing.T) {
	params := getConnectionManagerParams()
	params.PrivateKey = "./cert/sdjdjdks.pem"

	_, err := NewConnectionManager(params)

	if err == nil {
		t.Fatalf(`Expected error as private key does not exist`)
	}
}

func TestNewConnectionManagerCACertDoesNotExistAndVerify(t *testing.T) {
	params := getConnectionManagerParams()
	params.CaCert = "./cert/sdjdsjdksjdks.pem"
	params.SkipCertVerification = false

	_, err := NewConnectionManager(params)

	if err == nil {
		t.Fatal(`Expected error as ca cert does not exist and skip is false`)
	}
}

func TestNewConnectionManagerCACertDoesNotExistAndNotVerify(t *testing.T) {
	params := getConnectionManagerParams()
	params.CaCert = ""
	params.SkipCertVerification = true

	_, err := NewConnectionManager(params)

	if err != nil {
		t.Fatal(`an error should not be thrown`)
	}
}

func TestGetConnectionConnectionDoesNotExistAddsConnection(t *testing.T) {
	params := getConnectionManagerParams()

	m, _ := NewConnectionManager(params)

	conn, err := m.GetConnection("abc-123.com")

	if err != nil {
		t.Error(err)
	}

	if conn == nil {
		t.Fatal(`the connection should not be nil`)
	}

	conn2, _ := m.GetConnection("abc-123.com")

	if conn != conn2 {
		log.Fatalf(`should return the same connection instance`)
	}
}

func TestAddConnectionThrowsAnErrorIfFactoryThrowsError(t *testing.T) {
	params := getConnectionManagerParams()
	params.ConnFactory = func(clientConfig *tls.Config, server string) (PeerConnection, error) {
		return nil, errors.New("this is an error")
	}

	m, _ := NewConnectionManager(params)

	_, err := m.AddConnection("abc-123.com")

	if err == nil || err.Error() != "this is an error" {
		t.Error(err)
	}
}

func TestAddConnectionConnectionDoesNotExist(t *testing.T) {
	params := getConnectionManagerParams()

	m, _ := NewConnectionManager(params)

	conn, err := m.AddConnection("abc-123.com")

	if err != nil {
		t.Error(err)
	}

	if conn == nil {
		t.Fatal(`connection should not be nil`)
	}

	conn1, _ := m.GetConnection("abc-123.com")

	if conn != conn1 {
		t.Fatal(`underlying connections should be the same`)
	}
}

func TestHasConnectionConnectionDoesNotExist(t *testing.T) {
	params := getConnectionManagerParams()

	m, _ := NewConnectionManager(params)

	if m.HasConnection("abc-123.com") {
		t.Fatal(`should return that the connection does not exist`)
	}
}

func TestHasConnectionConnectionExists(t *testing.T) {
	params := getConnectionManagerParams()

	m, _ := NewConnectionManager(params)

	m.AddConnection("abc-123.com")

	if !m.HasConnection("abc-123.com") {
		t.Fatal(`should return that the connection exists`)
	}
}
