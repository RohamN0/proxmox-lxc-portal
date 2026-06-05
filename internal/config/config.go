package config

import "os"

type MikrotikConfig struct {
	Address            string
	User               string
	Password           string
	InsecureSkipVerify bool
}

type ProxmoxConfig struct {
	ApiUrl   string
	User     string
	Password string
}

func LoadProxmoxConfig() *ProxmoxConfig {
	apiUrl := os.Getenv("PROXMOX_API_URL")
	user := os.Getenv("PROXMOX_USER")
	password := os.Getenv("PROXMOX_PASSWORD")

	return &ProxmoxConfig{
		ApiUrl:   apiUrl,
		User:     user,
		Password: password,
	}

}

func LoadMikrotikConfig() *MikrotikConfig {

	address := os.Getenv("MIKROTIK_ADDRESS")
	user := os.Getenv("MIKROTIK_USER")
	password := os.Getenv("MIKROTIK_PASSWORD")
	insecureSkipVerify := os.Getenv("MIKROTIK_TLS_SKIP_VERIFY") == "true"

	return &MikrotikConfig{
		Address:            address,
		User:               user,
		Password:           password,
		InsecureSkipVerify: insecureSkipVerify,
	}
}
