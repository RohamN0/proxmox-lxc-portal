package router

import (
	"os"
	"proxmox-lxc-portal/internal/database"
	"proxmox-lxc-portal/internal/handlers"
	"proxmox-lxc-portal/internal/middleware"
	"proxmox-lxc-portal/internal/models"
	"proxmox-lxc-portal/internal/services"
	"strconv"
	"time"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	app.Get("/*", static.New("./internal/public"))

	// Middleware
	api := app.Group("/api", logger.New())

	// Auth
	userRepo := models.NewUserRepository(database.DB)
	refreshTokenRepo := models.NewRefreshTokenRepository(database.DB)
	jwtSecret := os.Getenv("SECRET")
	if jwtSecret == "" {
		panic("SECRET environment variable is required")
	}
	accessTTL := 15 * time.Minute
	if ttlEnv := os.Getenv("ACCESS_TOKEN_TTL_MINUTES"); ttlEnv != "" {
		if ttlMinutes, err := strconv.Atoi(ttlEnv); err == nil && ttlMinutes > 0 {
			accessTTL = time.Duration(ttlMinutes) * time.Minute
		}
	}
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtSecret, accessTTL)
	authHandler := handlers.NewAuthHandler(authService)

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	//auth.Post("/register", authHandler.Register)
	auth.Post("/logout", middleware.Protected(), authHandler.Logout)
	auth.Post("/refresh-token", authHandler.RefreshToken)

	// User
	userHandler := handlers.NewUserHandler(userRepo, refreshTokenRepo)
	user := api.Group("/users")
	user.Get("/:id", middleware.Protected(), userHandler.GetUser)
	user.Patch("/:id", middleware.Protected(), userHandler.UpdateUser)
	user.Delete("/:id", middleware.Protected(), userHandler.DeleteUser)

	// MikroTik
	mikrotikHandler, _ := handlers.NewMikroTikHandler()
	mikrotik := api.Group("/mikrotik")
	mikrotik.Post("/vlan", middleware.Protected(), mikrotikHandler.CreateVLAN)
	mikrotik.Post("/bridge/vlan", middleware.Protected(), mikrotikHandler.ConfigBridgeForVLAN)
	mikrotik.Post("/ip/assign", middleware.Protected(), mikrotikHandler.AssignIP)
	mikrotik.Post("/nat", middleware.Protected(), mikrotikHandler.ConfigureNAT)

	// Proxmox
	proxmoxHandler, err := handlers.NewProxmoxHandler()
	if err != nil {
    	log.Fatalf("Failed to initialize Proxmox handler: %v", err)
	}
	proxmox := api.Group("/proxmox")
	proxmox.Post("/lxc", middleware.Protected(), proxmoxHandler.CreateLXC)
	proxmox.Get("/resources", middleware.Protected(), proxmoxHandler.GetResources)

}
