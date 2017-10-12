package refnozzle

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
)

// NewClientMutualTLSConfig validates the provided key pair against a CA and
// returns a *tls.Config which permits any ExtKeyUsage.
func NewClientMutualTLSConfig(
	certFile string,
	keyFile string,
	caCertFile string,
	serverName string,
) (*tls.Config, error) {
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load keypair: %s", err)
	}

	certBytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
		return nil, errors.New("unable to load ca cert file")
	}

	certificate, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, err
	}

	verifyOptions := x509.VerifyOptions{
		Roots: caCertPool,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageAny,
		},
	}
	if _, err := certificate.Verify(verifyOptions); err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %s", err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		Certificates:       []tls.Certificate{tlsCert},
		ServerName:         serverName,
		RootCAs:            caCertPool,
	}

	return tlsConfig, err
}
