package services

import (
	"errors"
	"proxmox-lxc-portal/internal/models"
	"proxmox-lxc-portal/internal/repository"
)

type UserService struct {
	UserRepo         repository.UserRepositoryInterface
	RefreshTokenRepo repository.RefreshTokenRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface, refreshRepo repository.RefreshTokenRepositoryInterface) *UserService {
	return &UserService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshRepo,
	}
}

func (s *UserService) GetUser(userID uint) (*models.User, error) {
	user, err := s.UserRepo.GetUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateNames(userID uint, names string) (*models.User, error) {
	if names == "" {
		return nil, errors.New("names cannot be empty")
	}

	updatedUser, err := s.UserRepo.UpdateNames(userID, names)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return updatedUser, nil
}

func (s *UserService) DeleteUser(userID uint) error {
	err := s.RefreshTokenRepo.RevokeAllUserTokens(userID)
	if err != nil {
		return err
	}

	err = s.UserRepo.DeleteUser(userID)
	if err != nil {
		return ErrUserNotFound
	}

	return nil
}
