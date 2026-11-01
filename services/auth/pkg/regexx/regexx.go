package regexx

import (
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
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
