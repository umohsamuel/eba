package mfa

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	totpPeriod = 30
	totpSkew   = 1
)

var totpOpts = totp.ValidateOpts{
	Period:    totpPeriod,
	Skew:      0,
	Digits:    otp.DigitsSix,
	Algorithm: otp.AlgorithmSHA1,
}

type TOTPService struct {
	issuer string
}

func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{
		issuer: issuer,
	}
}

func (s *TOTPService) GenerateSecret(accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountName,
		Period:      totpPeriod,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
}

func (s *TOTPService) ValidateCode(secret, code string, now time.Time) (step int64, ok bool) {
	for offset := -totpSkew; offset <= totpSkew; offset++ {
		at := now.Add(time.Duration(offset*totpPeriod) * time.Second)
		valid, err := totp.ValidateCustom(code, secret, at, totpOpts)
		if err == nil && valid {
			return at.Unix() / totpPeriod, true
		}
	}

	return 0, false
}
