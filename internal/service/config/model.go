package config

import (
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type ConfigKey string

const (
	hostName           ConfigKey = "host.name"
	hostsTrusted                 = "hosts.trusted"
	devModeEnabled               = "dev_mode.enabled"
	httpAddrs                    = "http.addrs"
	tlsEnabled                   = "tls.enabled"
	sslCertificatePath           = "tls.ssl.certificate"
	sslPrivateKeyPath            = "tls.ssl.private_key"
)

type BootstrapCmd struct {
	HostName string
	Port     int
}

func (t BootstrapCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.HostName, v.Required, is.Domain),
		v.Field(&t.Port, v.Required, is.Port),
	)
}
