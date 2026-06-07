package repository

import (
	"proxmox-lxc-portal/internal/models"

	"gorm.io/gorm"
)

// UserRepositoryInterface defines the contract for User storage
type UserRepositoryInterface interface {
	CreateUser(string, string, string) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	GetUserByID(uint) (*models.User, error)
	DeleteUser(uint) error
	UpdateUser(uint, models.User) (*models.User, error)
	UpdateNames(uint, string) (*models.User, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

// CreateUser adds a new user to the database
func (r *UserRepository) CreateUser(email, username, passwordHash string) (*models.User, error) {
	user := &models.User{
		Email:    email,
		Username: username,
		Password: passwordHash,
	}

	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) DeleteUser(id uint) error {
	var user models.User
	if err := r.db.Where("id = ?", id).Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(id uint, updateUser models.User) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	user.Email = updateUser.Email
	user.Username = updateUser.Username
	user.Password = updateUser.Password
	user.Names = updateUser.Names

	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateNames updates the names field for a user and returns the updated user.
func (r *UserRepository) UpdateNames(id uint, names string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	user.Names = names
	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
