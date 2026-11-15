package mfa

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umohsamuel/eba/pkg/events"
	"github.com/umohsamuel/eba/services/auth/internal/domain/errors"
	"github.com/umohsamuel/eba/services/auth/internal/domain/mfa"
	"github.com/umohsamuel/eba/services/auth/internal/domain/outbox"
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

type SetupResult struct {
	OTPAuthURL string `json:"otpauth_url"`
	Secret     string `json:"secret"`
}

func (s *Service) Setup(ctx context.Context, userID uuid.UUID) (*SetupResult, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	key, err := s.totp.GenerateSecret(u.Email)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	encrypted, err := s.box.Encrypt([]byte(key.Secret()))
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	if err := s.repo.SavePendingTOTP(ctx, userID, encrypted); err != nil {
		return nil, err
	}

	return &SetupResult{OTPAuthURL: key.URL(), Secret: key.Secret()}, nil
}

func (s *Service) Confirm(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	t, err := s.repo.GetTOTP(ctx, userID)
	if err != nil {
		return nil, err
	}
	if t.Enabled() {
		return nil, errors.ErrMFAAlreadyEnabled
	}

	secret, err := s.box.Decrypt(t.SecretEncrypted)
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	step, ok := s.totp.ValidateCode(string(secret), strings.TrimSpace(code), time.Now())
	if !ok {
		return nil, errors.ErrMFACodeInvalid
	}

	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	if err := s.repo.ConfirmTOTP(ctx, userID, step, hashes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *Service) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.VerifyCode(ctx, userID, code); err != nil {
		return err
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	eventID := uuid.New()
	notifReq := events.NotificationRequested{
		EventID:    eventID.String(),
		Channel:    events.ChannelEmail,
		Template:   events.TemplateAuthMFADisabled,
		Recipient:  u.Email,
		OccurredAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(notifReq)
	if err != nil {
		return errors.ErrInternal.Wrap(err)
	}

	return s.repo.Disable(ctx, userID, outbox.OutboxEvent{
		ID:            eventID,
		AggregateType: events.TopicNotificationRequested,
		AggregateID:   notifReq.Recipient,
		Type:          notifReq.Template,
		Payload:       payload,
	})
}

func (s *Service) RegenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	if err := s.VerifyCode(ctx, userID, code); err != nil {
		return nil, err
	}

	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		return nil, errors.ErrInternal.Wrap(err)
	}

	if err := s.repo.ReplaceRecoveryCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}

	return codes, nil
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

const (
	recoveryCodeCount = 10
	recoveryCodeBytes = 5
)

func newRecoveryCodes() (codes []string, hashes []string, err error) {
	codes = make([]string, 0, recoveryCodeCount)
	hashes = make([]string, 0, recoveryCodeCount)

	for range recoveryCodeCount {
		b := make([]byte, recoveryCodeBytes)
		if _, err := rand.Read(b); err != nil {
			return nil, nil, err
		}

		raw := hex.EncodeToString(b)
		codes = append(codes, raw[:5]+"-"+raw[5:])
		hashes = append(hashes, hash.HashToken(normalizeRecoveryCode(raw)))
	}

	return codes, hashes, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(code))
}
