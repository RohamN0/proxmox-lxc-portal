package services

import "errors"

// Custom errors for the service layer
var (
	ErrUserNotFound = errors.New("user not found")
	ErrInternal     = errors.New("internal server error")
	ErrUnauthorized = errors.New("you are not authorized to perform this action")
	ErrNotFound     = errors.New("requested resource not found")
)
