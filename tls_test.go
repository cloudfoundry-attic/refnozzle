package refnozzle_test

import (
	"crypto/tls"

	"code.cloudfoundry.org/refnozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewClientMutualTLSConfig", func() {
	It("works with a valid key pair and CA", func() {
		c, err := refnozzle.NewClientMutualTLSConfig(
			"fixtures/refnozzle.crt",
			"fixtures/refnozzle.key",
			"fixtures/ca.crt",
			"refnozzle",
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(c.MinVersion).To(Equal(uint16(tls.VersionTLS12)))
		Expect(c.InsecureSkipVerify).To(Equal(false))
		Expect(c.Certificates).To(HaveLen(1))
		Expect(c.ServerName).To(Equal("refnozzle"))
		Expect(c.RootCAs).NotTo(BeNil())
	})

	It("errors on an invalid key pair", func() {
		_, err := refnozzle.NewClientMutualTLSConfig(
			"bad-cert",
			"bad-key",
			"fixtures/ca.crt",
			"refnozzle",
		)

		Expect(err).To(HaveOccurred())
	})

	It("errors on a bad CA", func() {
		_, err := refnozzle.NewClientMutualTLSConfig(
			"fixtures/refnozzle.crt",
			"fixtures/refnozzle.key",
			"bad-ca",
			"refnozzle",
		)

		Expect(err).To(HaveOccurred())
	})
})
