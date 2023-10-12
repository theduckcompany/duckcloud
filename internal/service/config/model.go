package config

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
