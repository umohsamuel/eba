package auth

import (
	"context"

	"github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	"github.com/umohsamuel/eba/services/auth/pkg/hash"
	"github.com/umohsamuel/eba/services/auth/pkg/regexx"
)

type AuthService struct {
	users user.UserRepository
}

func NewAuthService(users user.UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
) error {
	if !regexx.ValidateEmail(email) {
		return errors.ErrInvalidEmail
	}

	if !regexx.ValidatePassword(password) {
		return errors.ErrInvalidPassword
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return errors.ErrInternal
	}

	_, err = s.users.GetByEmail(ctx, email)
	if err == nil {
		return errors.ErrEmailAlreadyExists
	}

	newUser := &user.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	err = s.users.Create(ctx, newUser)
	if err != nil {
		return errors.ErrInternal
	}

	return nil
}

// A dummy hash to run against if a user isn't found to mitigate timing attacks.
const dummyHash = "$2a$10$xK9vG/5F7y8Kz7v9B6h2u.D8r1e2f3g4h5i6j7k8l9m0n1o2p3q4r"

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) error {
	if !regexx.ValidateEmail(email) {
		return errors.ErrInvalidEmail
	}

	if !regexx.ValidatePassword(password) {
		return errors.ErrInvalidPassword
	}

	u, err := s.users.GetByEmail(ctx, email)

	var targetHash string
	if err != nil {
		targetHash = dummyHash
	} else {
		targetHash = u.PasswordHash
	}

	isCorrectPassword := hash.CheckPasswordHash(password, targetHash)
	if !isCorrectPassword {
		return errors.ErrInvalidCredentials
	}

	if err != nil || !isCorrectPassword {
		return errors.ErrInvalidCredentials
	}

	return nil
}
