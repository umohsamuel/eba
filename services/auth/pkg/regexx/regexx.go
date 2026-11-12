package regexx

import (
	"net/mail"
	"regexp"
	"strings"
)

const maxEmailLength = 254

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) bool {
	if len(email) > maxEmailLength {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}

	domain := email[strings.LastIndex(email, "@")+1:]
	return strings.Contains(domain, ".") && !strings.HasSuffix(domain, ".")
}

var (
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	digitRegex     = regexp.MustCompile(`[0-9]`)
	specialRegex   = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};':",\./<>\?~|]`)
)

func ValidatePassword(password string) bool {
	if len(password) < 8 || len(password) > 32 {
		return false
	}

	if !lowercaseRegex.MatchString(password) {
		return false
	}
	if !uppercaseRegex.MatchString(password) {

		return false
	}

	if !digitRegex.MatchString(password) {
		return false
	}

	if !specialRegex.MatchString(password) {
		return false
	}

	return true
}
