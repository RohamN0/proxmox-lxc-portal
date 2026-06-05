// internal/handlers/mikrotik.go
package handlers

import (
	"crypto/tls"
	"fmt"
	"net"
	"proxmox-lxc-portal/internal/config"
	"strconv"
	"strings"

	"github.com/go-routeros/routeros"
	"github.com/gofiber/fiber/v3"
)

// MikroTikHandler holds the MikroTik API client and configuration
type MikroTikHandler struct {
	Address   string
	User      string
	Password  string
	TLSConfig *tls.Config
}

// Request/Response structs (kept from your original code)
type VLANInterface struct {
	Name      string `json:"name"`
	VLANID    int    `json:"vlan_id"`
	Interface string `json:"interface"`
}

type BridgeForVLAN struct {
	Bridge    string `json:"bridge"`
	VLANID    int    `json:"vlan_id"`
	Interface string `json:"interface"`
	Tagged    string `json:"tagged"`
	UnTagged  string `json:"untagged"`
}

type IP struct {
	Address   string `json:"address"`
	Interface string `json:"interface"`
}

type NAT struct {
	Chain      string `json:"chain"`
	DstAddress string `json:"dst-address"`
	Protocol   string `json:"protocol"`
	DstPort    int    `json:"dst-port"`
	Action     string `json:"action"`
	ToAddress  string `json:"to-address"`
	ToPort     int    `json:"to-port"`
}

// NewMikroTikHandler creates a new handler instance with config from env vars
func NewMikroTikHandler() (*MikroTikHandler, error) {
	mikrotikConfig := config.LoadMikrotikConfig()

	if mikrotikConfig.Address == "" || mikrotikConfig.User == "" || mikrotikConfig.Password == "" {
		return nil, fmt.Errorf("MIKROTIK_ADDRESS, MIKROTIK_USER, and MIKROTIK_PASSWORD must be set")
	}

	// Optional: customize TLS config if needed
	tlsConfig := &tls.Config{
		InsecureSkipVerify: mikrotikConfig.InsecureSkipVerify,
	}

	return &MikroTikHandler{
		Address:   mikrotikConfig.Address,
		User:      mikrotikConfig.User,
		Password:  mikrotikConfig.Password,
		TLSConfig: tlsConfig,
	}, nil
}

// connect creates a new MikroTik API connection (called per request for safety)
func (h *MikroTikHandler) connect() (*routeros.Client, error) {
	client, err := routeros.DialTLS(h.Address, h.User, h.Password, h.TLSConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MikroTik: %w", err)
	}
	return client, nil
}

// CreateVLAN handles POST /api/v1/mikrotik/vlan
func (h *MikroTikHandler) CreateVLAN(c fiber.Ctx) error {
	var req VLANInterface
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Interface == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The interface field should not be empty",
		})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The name field should not be empty",
		})
	}
	if req.VLANID < 1 || req.VLANID > 4094 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "VLAN ID must be between 1 and 4094",
		})
	}

	client, err := h.connect()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer client.Close()

	// Execute /interface/vlan/add with correct RouterOS syntax
	reply, err := client.Run(
		"/interface/vlan/add",
		"=name="+req.Name,
		"=vlan-id="+strconv.Itoa(req.VLANID),
		"=interface="+req.Interface,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create VLAN: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "VLAN interface created",
		"detail":  reply,
	})
}

// ConfigBridgeForVLAN handles POST /api/v1/mikrotik/bridge/vlan
func (h *MikroTikHandler) ConfigBridgeForVLAN(c fiber.Ctx) error {
	var req BridgeForVLAN
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Interface == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The interface field should not be empty",
		})
	}
	if req.Bridge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The bridge field should not be empty",
		})
	}
	if req.VLANID < 1 || req.VLANID > 4094 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "VLAN ID must be between 1 and 4094",
		})
	}
	if req.Tagged == "" && req.UnTagged == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one of tagged or untagged ports must be specified",
		})
	}

	client, err := h.connect()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer client.Close()

	// Add bridge port with PVID for untagged traffic
	_, err = client.Run(
		"/interface/bridge/port/add",
		"=bridge="+req.Bridge,
		"=interface="+req.Interface,
		"=pvid="+strconv.Itoa(req.VLANID),
		"=frame-types=admit-only-vlan-tagged",
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add bridge port: " + err.Error(),
		})
	}

	// Add VLAN to bridge with tagged/untagged ports
	vlanCmd := []string{
		"/interface/bridge/vlan/add",
		"=bridge=" + req.Bridge,
		"=vlan-ids=" + strconv.Itoa(req.VLANID),
	}
	if req.Tagged != "" {
		vlanCmd = append(vlanCmd, "=tagged="+req.Tagged)
	}
	if req.UnTagged != "" {
		vlanCmd = append(vlanCmd, "=untagged="+req.UnTagged)
	}

	_, err = client.Run(vlanCmd...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to configure bridge VLAN: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Bridge configured for VLAN",
		"vlan_id": req.VLANID,
	})
}

// AssignIP handles POST /api/v1/mikrotik/ip/assign
func (h *MikroTikHandler) AssignIP(c fiber.Ctx) error {
	var req IP
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Interface == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "The interface field should not be empty",
		})
	}
	// ParseIP returns nil for invalid IPs, so check for nil
	if net.ParseIP(strings.Split(req.Address, "/")[0]) == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid IP address format (use CIDR notation, e.g., 192.168.1.1/24)",
		})
	}

	client, err := h.connect()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer client.Close()

	_, err = client.Run(
		"/ip/address/add",
		"=address="+req.Address,
		"=interface="+req.Interface,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign IP: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "IP address assigned",
		"address":   req.Address,
		"interface": req.Interface,
	})
}

// ConfigureNAT handles POST /api/v1/mikrotik/nat
func (h *MikroTikHandler) ConfigureNAT(c fiber.Ctx) error {
	var req NAT
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate chain (FIXED: use AND logic, not OR)
	if req.Chain != "dstnat" && req.Chain != "srcnat" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Chain must be 'dstnat' or 'srcnat'",
		})
	}

	// Validate destination IP
	if net.ParseIP(strings.Split(req.DstAddress, "/")[0]) == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid destination IP address",
		})
	}

	// Validate protocol (allow empty or common protocols)
	validProtocols := map[string]bool{"": true, "tcp": true, "udp": true, "icmp": true, "all": true}
	if !validProtocols[req.Protocol] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Protocol must be tcp, udp, icmp, all, or empty",
		})
	}

	// Validate ports
	if req.DstPort < 0 || req.DstPort > 65535 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Destination port must be between 0 and 65535",
		})
	}
	if req.ToPort < 0 || req.ToPort > 65535 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Target port must be between 0 and 65535",
		})
	}

	// Validate action (common NAT actions)
	validActions := map[string]bool{
		"accept": true, "drop": true, "reject": true,
		"dst-nat": true, "src-nat": true, "masquerade": true,
	}
	if !validActions[req.Action] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid action. Must be: accept, drop, reject, dst-nat, src-nat, or masquerade",
		})
	}

	// Validate to-address for NAT actions that require it
	if (req.Action == "dst-nat" || req.Action == "src-nat") && net.ParseIP(req.ToAddress) == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid to-address IP for NAT action",
		})
	}

	client, err := h.connect()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer client.Close()

	// Build NAT command dynamically (only include non-empty fields)
	natCmd := []string{"/ip/firewall/nat/add", "=chain=" + req.Chain}

	if req.DstAddress != "" {
		natCmd = append(natCmd, "=dst-address="+req.DstAddress)
	}
	if req.Protocol != "" {
		natCmd = append(natCmd, "=protocol="+req.Protocol)
	}
	if req.DstPort > 0 {
		natCmd = append(natCmd, "=dst-port="+strconv.Itoa(req.DstPort))
	}
	if req.Action != "" {
		natCmd = append(natCmd, "=action="+req.Action)
	}
	if req.ToAddress != "" {
		natCmd = append(natCmd, "=to-addresses="+req.ToAddress) // Note: RouterOS uses "to-addresses"
	}
	if req.ToPort > 0 {
		natCmd = append(natCmd, "=to-ports="+strconv.Itoa(req.ToPort))
	}

	reply, err := client.Run(natCmd...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create NAT rule: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "NAT rule configured successfully",
		"chain":   req.Chain,
		"action":  req.Action,
		"detail":  reply,
	})
}
