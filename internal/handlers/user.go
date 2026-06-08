package handlers

import (
	"errors"
	"proxmox-lxc-portal/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// UserHandler contains HTTP handlers for users.
type UserHandler struct {
	UserService *services.UserService
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func validToken(t *jwt.Token, id string) bool {
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return false
	}
	return sub == id
}

func parseUserID(c fiber.Ctx) (uint, string, error) {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return 0, "", errors.New("invalid user ID")
	}
	return uint(id), idParam, nil
}

// GetUser get a user
func (uh *UserHandler) GetUser(c fiber.Ctx) error {
	id, idString, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	tok, ok := c.Locals("user").(*jwt.Token)
	if !ok || !validToken(tok, idString) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token id",
			"data":    nil,
		})
	}

	user, err := uh.UserService.GetUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User found",
		"data":    user,
	})
}

// UpdateUser update user
func (uh *UserHandler) UpdateUser(c fiber.Ctx) error {
	var input struct {
		Names string `json:"names"`
	}
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"data":    err.Error(),
		})
	}

	id, idString, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	tok, ok := c.Locals("user").(*jwt.Token)
	if !ok || !validToken(tok, idString) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token id",
			"data":    nil,
		})
	}

	updatedUser, err := uh.UserService.UpdateNames(id, input.Names)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "User update failed",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User successfully updated",
		"data":    updatedUser,
	})
}

// DeleteUser delete user
func (uh *UserHandler) DeleteUser(c fiber.Ctx) error {
	id, idString, err := parseUserID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	tok, ok := c.Locals("user").(*jwt.Token)
	if !ok || !validToken(tok, idString) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid token id",
			"data":    nil,
		})
	}

	err = uh.UserService.DeleteUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "User not found",
			"data":    nil,
		})
	}

	c.Locals("user", nil)
	c.ClearCookie("jwt")

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User successfully deleted",
		"data":    nil,
	})
}
