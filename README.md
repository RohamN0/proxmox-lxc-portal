# Proxmox LXC Portal

A backend web portal and API for Proxmox VE, designed to manage users and automate the deployment of LXC containers with integrated network configuration (MikroTik).

This project provides a RESTful API to handle user authentication, LXC container creation, VLAN management, and IP assignment, acting as a self-service control panel for virtualized environments.

## ✨ Features

- **User Authentication**: Secure JWT-based authentication system with access and refresh tokens.
- **User Management**: Endpoints to get, update, and delete user profiles.
- **Proxmox Integration**:
  - Create LXC containers with a single API call, specifying resources (RAM, CPU cores) and network configuration (IP address, VLAN ID).
  - Fetch node resource usage (CPU, RAM, Storage) to monitor the health of your Proxmox host.
- **Network Automation**:
  - Create and manage VLANs on a MikroTik router.
  - Configure bridge interfaces for VLANs.
  - Assign IP addresses and configure NAT rules.

## 🛠️ Tech Stack

- **Language**: Go
- **Web Framework**: [Fiber](https://gofiber.io/) for high-performance routing and middleware.
- **Database**: SQLite with [GORM](https://gorm.io/) as the ORM for data persistence.
- **Proxmox Client**: [`github.com/Telmate/proxmox-api-go`](https://github.com/Telmate/proxmox-api-go) to interface with the Proxmox VE API.
- **Authentication**: JWT (JSON Web Tokens) for secure API access.

## 📋 Prerequisites

- A running **Proxmox VE** instance.
- A **MikroTik** router for network automation (optional, but required for VLAN/IP features).
- A **Linux/macOS/Windows** server to host this application.
- **Go** 1.18 or later (only if building from source).

## 🚀 Getting Started

### 1. Installation

Clone the repository:

```bash
git clone https://github.com/RohamN0/proxmox-lxc-portal.git
cd proxmox-lxc-portal
```

Build the application:

```bash
go build -o portal cmd/main.go
```

### 2. Configuration

Create a `.env` file in the root directory with the following variables:

```env
# JWT Configuration
SECRET=your-very-secure-jwt-secret
ACCESS_TOKEN_TTL_MINUTES=15

# Proxmox Configuration
PROXMOX_API_URL=https://your-proxmox-host:8006/api2/json
PROXMOX_USER=your-api-user@pam
PROXMOX_PASSWORD=your-api-password

# MikroTik Configuration (Optional)
MIKROTIK_API_URL=https://your-mikrotik-ip/rest
MIKROTIK_USER=your-api-username
MIKROTIK_PASSWORD=your-api-password
```

### 3. Run the Application

```bash
./portal
```

The server will start on `http://localhost:3000`. The main UI will be available at `http://localhost:3000/`.

## 📚 API Endpoints

All endpoints under `/api` are protected by JWT authentication, except for the login/refresh endpoints.

| Method | Endpoint               | Description                           | Authentication |
| :----- | :--------------------- | :------------------------------------ | :------------- |
| POST   | `/api/auth/login`      | Login and receive access token        | None           |
| POST   | `/api/auth/logout`     | Invalidate the current session        | JWT Token      |
| POST   | `/api/auth/refresh-token` | Obtain a new access token           | Refresh Token  |
| GET    | `/api/users/:id`       | Get user details                      | JWT Token      |
| PATCH  | `/api/users/:id`       | Update user details                   | JWT Token      |
| DELETE | `/api/users/:id`       | Delete a user                         | JWT Token      |
| POST   | `/api/proxmox/lxc`     | Create a new LXC container            | JWT Token      |
| GET    | `/api/proxmox/resources` | Get Proxmox node resource usage    | JWT Token      |
| POST   | `/api/mikrotik/vlan`   | Create a VLAN on MikroTik             | JWT Token      |
| POST   | `/api/mikrotik/bridge/vlan` | Configure bridge for a VLAN      | JWT Token      |
| POST   | `/api/mikrotik/ip/assign` | Assign an IP address               | JWT Token      |
| POST   | `/api/mikrotik/nat`    | Configure NAT rules on MikroTik       | JWT Token      |

### Example: Create LXC Container

```bash
POST /api/proxmox/lxc
Content-Type: application/json
Authorization: Bearer <your_access_token>

{
  "node": "server01",
  "vmid": 101,
  "hostname": "my-container",
  "template": "ubuntu-22.04-standard",
  "storage": "data-hdd-01",
  "ram": 2048,
  "cores": 2,
  "password": "container-root-password",
  "ip": "192.168.1.100/24",
  "vlan-id": 10
}
```

## 🖥️ Basic UI

A simple static HTML interface is provided at the root path, allowing you to log in and submit requests to the API. This is intended for demonstration purposes only.