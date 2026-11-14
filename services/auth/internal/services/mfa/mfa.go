package totp

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPService struct {
	issuer string
}

func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{
		issuer: issuer,
	}
}

func (s *TOTPService) GenerateSecret(username string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      6,
		Period:      30,
		// You can generate secret manually and set it here.
		// If secret is not provided, library will generate a random secret
		// Secret: secret,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *TOTPService) ValidateCode(secret string, code string) bool {
	valid := totp.Validate(code, secret)
	return valid
}
