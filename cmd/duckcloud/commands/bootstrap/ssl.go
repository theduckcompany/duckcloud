package bootstrap

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/theduckcompany/duckcloud/internal/service/config"
)

func enableSSL(cmd *cobra.Command, configSvc config.Service) {
	err := configSvc.EnableTLS(cmd.Context())
	if err != nil {
		printErrAndExit(cmd, err)
	}
}

func disableSSL(cmd *cobra.Command, configSvc config.Service) {
	err := configSvc.DisableTLS(cmd.Context())
	if err != nil {
		printErrAndExit(cmd, err)
	}
}

func setupSSLCertificate(cmd *cobra.Command, configSvc config.Service, folderPath string) {
	const (
		SelfSignedCertif   = "Generate a self-signed certificate"
		UserProvidedCertif = "I already have my certificate"
	)

	sslEnable, err := configSvc.IsTLSEnabled(cmd.Context())
	switch {
	case err == nil:
		cmd.Printf("SSL enabled: %v\n", sslEnable)
	case errors.Is(err, config.ErrNotInitialized):
		// continue
	default:
		printErrAndExit(cmd, fmt.Errorf("failed to check if TLS is enabled: %w", err))
	}

	if !sslEnable {
		return
	}

	certifPath, privateKeyPath, err := configSvc.GetSSLPaths(cmd.Context())
	if err != nil && !errors.Is(err, config.ErrNotInitialized) {
		printErrAndExit(cmd, fmt.Errorf("failed to retrieve the SSL files paths: %w", err))
	}

	if certifPath != "" && privateKeyPath != "" {
		cmd.Printf("SSL certificate path: %s\n", certifPath)
		cmd.Printf("SSL private key path: %s\n", privateKeyPath)
		return
	}

	devMode, err := configSvc.IsDevModeEnabled(cmd.Context())
	if err != nil {
		printErrAndExit(cmd, err)
	}

	if devMode {
		generateSelfSignedCertificate(cmd, configSvc, folderPath)
		return
	}

	prompt := &survey.Select{
		Message: `What kind of SSL certificate do you want to setup?`,
		Options: []string{SelfSignedCertif, UserProvidedCertif},
	}

	var res string
	err = survey.AskOne(prompt, &res)
	if err != nil {
		printErrAndExit(cmd, err)
	}

	switch res {
	case SelfSignedCertif:
		generateSelfSignedCertificate(cmd, configSvc, folderPath)
	case UserProvidedCertif:
		certifPath := askPath(cmd, "Certificate path (.pem)", false)
		privateKeyPath := askPath(cmd, "PrivateKey path (.pem)", false)
		err = configSvc.SetSSLPaths(cmd.Context(), certifPath, privateKeyPath)
		if err != nil {
			printErrAndExit(cmd, err)
		}
	default:
		printErrAndExit(cmd, errors.New("invalid selection"))
	}
}

func generateSelfSignedCertificate(cmd *cobra.Command, confiSvc config.Service, folderPath string) {
	sslFolder := path.Join(folderPath, "ssl")
	certificatePath := path.Join(sslFolder, "cert.pem")
	privateKeyPath := path.Join(sslFolder, "key.pem")

	err := os.MkdirAll(sslFolder, 0o700)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to create the SSL folder: %w", err))
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to GenerateKey: %w", err))
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to generate serial number: %w", err))
	}

	hostname, err := confiSvc.GetHostName(cmd.Context())
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to fetch the hostname: %w", err))
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Duck Corp"},
		},
		DNSNames:  []string{hostname},
		NotBefore: time.Now(),
		// NotAfter:  time.Now().Add(3 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to create certificate: %v", err))
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		printErrAndExit(cmd, errors.New("failed to encode certificate to PEM"))
	}

	if err := os.WriteFile(certificatePath, pemCert, 0o644); err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to write the certificate into the data folder: %w", err))
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		printErrAndExit(cmd, fmt.Errorf("unable to marshal private key: %v", err))
	}

	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		printErrAndExit(cmd, errors.New("failed to encode key to PEM"))
	}

	if err := os.WriteFile(privateKeyPath, pemKey, 0o600); err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to write the certificate into the data folder: %w", err))
	}

	if err := confiSvc.SetSSLPaths(cmd.Context(), certificatePath, privateKeyPath); err != nil {
		printErrAndExit(cmd, fmt.Errorf("failed to save the SSL config in db: %w", err))
	}

	cmd.Printf("Certificate setup: %q\n", certificatePath)
	cmd.Printf("Private Key setup: %q\n", privateKeyPath)
}

func askPath(cmd *cobra.Command, message string, expectDir bool) string {
	prompt := survey.Input{
		Message: message,
		Suggest: func(toComplete string) []string {
			res := []string{}
			entries, _ := os.ReadDir(toComplete)
			if entries != nil {
				for _, entry := range entries {
					res = append(res, path.Join(toComplete, entry.Name()))
				}
			}
			return res
		},
	}

	var path string
	survey.AskOne(&prompt, &path, survey.WithValidator(func(input interface{}) error {
		inputStr, ok := input.(string)
		if !ok {
			return errors.New("must be a string")
		}

		info, err := os.Stat(inputStr)
		if err != nil {
			return err
		}

		if info.IsDir() != expectDir {
			if expectDir {
				return errors.New("expect a directory")
			}

			return errors.New("expect a file")
		}

		return nil
	}))

	return path
}
