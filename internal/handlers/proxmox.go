package handlers

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/gofiber/fiber/v3"
)

// ProxmoxHandler holds the authenticated Proxmox client
type ProxmoxHandler struct {
	Client *proxmox.Client
}

// LXCRequest represents the JSON payload for creating an LXC container
type LXCRequest struct {
	Node     string `json:"node"`
	VMID     int    `json:"vmid"`
	Hostname string `json:"hostname"`
	Template string `json:"template"`
	Storage  string `json:"storage"`
	RAM      int    `json:"ram"`
	Cores    int    `json:"cores"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	VLANID   int    `json:"vlan-id"`
}

// NewProxmoxHandler creates a handler by reading credentials from environment variables
func NewProxmoxHandler() (*ProxmoxHandler, error) {
	apiUrl := os.Getenv("PROXMOX_API_URL")
	user := os.Getenv("PROXMOX_USER")
	password := os.Getenv("PROXMOX_PASSWORD")

	if apiUrl == "" || user == "" || password == "" {
		return nil, fmt.Errorf("PROXMOX_API_URL, PROXMOX_USER and PROXMOX_PASSWORD must be set")
	}

	// In production, use proper TLS certificates
	tlsconf := &tls.Config{InsecureSkipVerify: true}

	// NewClient signature (latest): (url, httpClient, token, tlsConfig, proxy, timeout, debug)
	// We pass nil httpClient (will use default), empty token (we'll login with user/password), empty proxy
	client, err := proxmox.NewClient(apiUrl, nil, "", tlsconf, "", 300, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create Proxmox client: %w", err)
	}

	// Login now requires a context (use background)
	if err := client.Login(context.Background(), user, password, ""); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &ProxmoxHandler{Client: client}, nil
}

// GetResources returns node and storage resource usage
func (p *ProxmoxHandler) GetResources(c fiber.Ctx) error {
	ctx := context.Background()
	nodeName := c.Query("node", "server01")
	storageName := c.Query("storage", "data-hdd-01")

	// CPU & RAM – GetItemConfigMapStringInterface now takes (ctx, url, key, extra)
	// The extra string can be empty; the method's signature in latest is (ctx, url, key, retry string)
	nodeInfo, err := p.Client.GetItemConfigMapStringInterface(ctx,
		"/nodes/"+nodeName+"/status", "", "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get node status: " + err.Error(),
		})
	}

	memory := nodeInfo["memory"].(map[string]interface{})
	ramTotal := memory["total"].(float64)
	ramUsed := memory["used"].(float64)

	cpuCores := nodeInfo["cpuinfo"].(map[string]interface{})["cpus"].(float64)
	cpuUsage := nodeInfo["cpu"].(float64)

	// Storage
	storageInfo, err := p.Client.GetItemConfigMapStringInterface(ctx,
		fmt.Sprintf("/nodes/%s/storage/%s/status", nodeName, storageName), "", "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get storage status: " + err.Error(),
		})
	}

	diskTotal := storageInfo["total"].(float64)
	diskUsed := storageInfo["used"].(float64)

	return c.JSON(fiber.Map{
		"node": nodeName,
		"cpu": fiber.Map{
			"cores":      cpuCores,
			"usage_perc": cpuUsage * 100,
		},
		"ram_gb": fiber.Map{
			"total":     ramTotal / (1024 * 1024 * 1024),
			"available": (ramTotal - ramUsed) / (1024 * 1024 * 1024),
		},
		"storage_gb": fiber.Map{
			"name":      storageName,
			"total":     diskTotal / (1024 * 1024 * 1024),
			"available": (diskTotal - diskUsed) / (1024 * 1024 * 1024),
		},
	})
}

// CreateLXC creates a new LXC container with the provided configuration
func (p *ProxmoxHandler) CreateLXC(c fiber.Ctx) error {
	req := new(LXCRequest)
	if err := c.Bind().Body(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation
	if req.VMID == 0 || req.Node == "" || req.Template == "" || req.Storage == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields: vmid, node, template, storage",
		})
	}

	config := proxmox.ConfigLxc{
		Ostemplate: req.Template,
		Arch:       "amd64",
		Cores:      req.Cores,
		Memory:     req.RAM,
		Hostname:   req.Hostname,
		Password:   req.Password,
		OnBoot:       true,
		// Networks field must be a QemuDevices map, not a slice
		Networks: proxmox.QemuDevices{
			0: map[string]interface{}{
				"name":     "eth0",
				"bridge":   "vmbr0",
				"ip":       req.IP,
				"tag": req.VLANID,
			},
		},
		RootFs: map[string]interface{}{
			"storage": req.Storage,
			"size":    "8G",
		},
	}

	// CreateLxc now requires a context
	vmRef := proxmox.NewVmRef(proxmox.GuestID(req.VMID))
	vmRef.SetNode(req.Node)

	err := config.CreateLxc(context.Background(), vmRef, p.Client)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to create LXC: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message":  "LXC created successfully",
		"vmid":     req.VMID,
		"hostname": req.Hostname,
	})
}
