package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	gLog "github.com/gogf/gf/v2/frame/g"
)

type SSLGenerator interface {
	Generate(ctx context.Context, outputDir string, commonName string, dnsNames []string, ipAddresses []string) error
}

type selfSignedSSLGenerator struct {
	org      string
	validFor time.Duration
	bits     int
}

func NewSSLGenerator() SSLGenerator {
	return &selfSignedSSLGenerator{
		org:      "OPSKVM",
		validFor: 365 * 24 * time.Hour,
		bits:     2048,
	}
}

func (g *selfSignedSSLGenerator) Generate(ctx context.Context, outputDir string, commonName string, dnsNames []string, ipAddresses []string) error {
	certFile := filepath.Join(outputDir, "server.crt")
	keyFile := filepath.Join(outputDir, "server.key")

	if _, err := os.Stat(certFile); err == nil {
		gLog.Log().Infof(ctx, "SSL certificate already exists, skipping generation: %s", certFile)
		return nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, g.bits)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := &privateKey.PublicKey

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{g.org},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(g.validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
	}
	for _, ipStr := range ipAddresses {
		if ip := net.ParseIP(ipStr); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	keyOut.Close()

	gLog.Log().Infof(ctx, "SSL certificate generated successfully in: %s", outputDir)
	gLog.Log().Infof(ctx, "  Certificate: %s", certFile)
	gLog.Log().Infof(ctx, "  Private Key: %s", keyFile)

	return nil
}
