package repository

import (
	"proxmox-lxc-portal/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshTokenRepositoryInterface defines the contract for token storage
type RefreshTokenRepositoryInterface interface {
	CreateRefreshToken(userId uint, ttl time.Duration) (*models.RefreshToken, error)
	GetRefreshToken(tokenString string) (*models.RefreshToken, error)
	RevokeRefreshToken(tokenString string) error
	RevokeAllUserTokens(userId uint) error
}

// RefreshTokenRepository handles database operations for refresh tokens
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepositoryInterface {
	return &RefreshTokenRepository{db: db}
}

// CreateRefreshToken creates a new refresh token for a user
func (r *RefreshTokenRepository) CreateRefreshToken(userId uint, ttl time.Duration) (*models.RefreshToken, error) {
	token := &models.RefreshToken{
		Token:     uuid.New().String(),
		UserId:    userId,
		ExpiresAt: time.Now().Add(ttl),
		Revoked:   false,
	}
	if err := r.db.Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

// GetRefreshToken retrieves a refresh token by its token string
func (r *RefreshTokenRepository) GetRefreshToken(tokenString string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token = ?", tokenString).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *RefreshTokenRepository) RevokeRefreshToken(tokenString string) error {
	if err := r.db.Model(&models.RefreshToken{}).Where("token = ?", tokenString).Update("revoked", true).Error; err != nil {
		return err
	}
	return nil
}

// RevokeAllUserTokens marks all refresh tokens for a user as revoked.
func (r *RefreshTokenRepository) RevokeAllUserTokens(userId uint) error {
	if err := r.db.Model(&models.RefreshToken{}).Where("user_id = ?", userId).Update("revoked", true).Error; err != nil {
		return err
	}
	return nil
}
