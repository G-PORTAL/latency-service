package config

import (
	"os"
	"path"
	"strings"
)

type Config struct {
	Hostname           string
	DataDirectory      string
	ListenAddressHTTP  string
	ListenAddressHTTPS string
	LetsEncrypt        bool
	Logging            bool
	LetsEncryptEmail   string
	CertPath           string
	KeyPath            string
}

func (c Config) GetCertificateDirectory() string {
	return path.Join(c.DataDirectory, "certs")
}

func (c Config) GetLogFile() string {
	return path.Join(c.DataDirectory, "latency.logs")
}

func GetConfig() Config {
	return Config{
		Hostname:           getEnv("LATENCY_HOST", "example-region.my-domain.com"),
		DataDirectory:      getEnv("LATENCY_DATA_DIRECTORY", "/data"),
		ListenAddressHTTP:  getEnv("LATENCY_LISTEN_HTTP", "0.0.0.0:8080"),
		ListenAddressHTTPS: getEnv("LATENCY_LISTEN_HTTPS", "0.0.0.0:8443"),
		CertPath:           getEnv("CERT_PATH", ""),
		KeyPath:            getEnv("KEY_PATH", ""),
		LetsEncrypt:        getEnvBool("LATENCY_LETS_ENCRYPT", false),
		Logging:            getEnvBool("LATENCY_LOGGING", true),
		LetsEncryptEmail:   getEnv("LATENCY_LETS_ENCRYPT_EMAIL", ""),
	}
}

func getEnv(key string, def string) string {
	envValue := os.Getenv(key)
	if envValue == "" {
		return def
	}

	return envValue
}

func getEnvBool(key string, def bool) bool {
	envValue := os.Getenv(key)
	if strings.ToLower(envValue) == "true" {
		return true
	}
	if strings.ToLower(envValue) == "false" {
		return false
	}

	return def
}
