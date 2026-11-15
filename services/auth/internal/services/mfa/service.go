package mfa

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/mfa"
	"github.com/umohsamuel/eba/services/auth/internal/domain/user"
	"github.com/umohsamuel/eba/services/auth/pkg/hash"
	"github.com/umohsamuel/eba/services/auth/pkg/secretbox"
)

type Service struct {
	repo  mfa.MFARepository
	users user.UserRepository
	box   secretbox.SecretBox
	totp  *TOTPService
}

type Config struct {
	Repository mfa.MFARepository
	Users      user.UserRepository
	SecretBox  secretbox.SecretBox
	Issuer     string
}

func New(deps Config) *Service {
	return &Service{
		repo:  deps.Repository,
		users: deps.Users,
		box:   deps.SecretBox,
		totp:  NewTOTPService(deps.Issuer),
	}
}

func (s *Service) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.repo.IsEnabled(ctx, userID)
}

func (s *Service) VerifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.ErrMFACodeInvalid
	}

	t, err := s.repo.GetTOTP(ctx, userID)
	if err != nil {
		return err
	}
	if !t.Enabled() {
		return errors.ErrMFANotEnrolled
	}

	secret, err := s.box.Decrypt(t.SecretEncrypted)
	if err != nil {
		return errors.ErrInternal.Wrap(err)
	}

	if step, ok := s.totp.ValidateCode(string(secret), code, time.Now()); ok {
		return s.repo.AdvanceStep(ctx, userID, step)
	}

	return s.repo.ConsumeRecoveryCode(ctx, userID, hash.HashToken(normalizeRecoveryCode(code)))
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(code))
}
