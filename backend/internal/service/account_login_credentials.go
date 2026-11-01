package service

import (
	"encoding/base32"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const AccountLoginUsernameKey = "login_username"
const AccountLoginPasswordKey = "login_password"
const AccountLoginTOTPKey = "login_totp_secret"

// Login records are administrator notes, never upstream authentication inputs.
func NormalizeAccountLoginCredentials(credentials map[string]any) error {
	for _, key := range []string{AccountLoginUsernameKey, AccountLoginPasswordKey, AccountLoginTOTPKey} {
		raw, exists := credentials[key]
		if !exists {
			continue
		}
		value, ok := raw.(string)
		if !ok || len(value) > 4096 {
			return apperrors.BadRequest("INVALID_ACCOUNT_LOGIN", "Login credentials must be strings of at most 4096 bytes")
		}
		if key == AccountLoginTOTPKey && strings.TrimSpace(value) == "" {
			credentials[key] = ""
		} else if key == AccountLoginTOTPKey {
			normalized, err := normalizeAccountTOTP(value)
			if err != nil {
				return err
			}
			credentials[key] = normalized
		}
	}
	return nil
}

func normalizeAccountTOTP(value string) (string, error) {
	invalid := func() (string, error) {
		return "", apperrors.BadRequest("INVALID_ACCOUNT_TOTP", "Enter a valid Base32 secret or TOTP otpauth URI (SHA1/SHA256/SHA512, 6/8 digits, period 1-300 seconds)")
	}
	value = strings.TrimSpace(value)
	secret := value
	var uri *url.URL
	if strings.Contains(value, "://") {
		var err error
		uri, err = url.Parse(value)
		if err != nil || uri.Scheme != "otpauth" || uri.Host != "totp" || uri.User != nil {
			return invalid()
		}
		q, err := url.ParseQuery(uri.RawQuery)
		if err != nil {
			return invalid()
		}
		for _, key := range []string{"secret", "algorithm", "digits", "period"} {
			if len(q[key]) > 1 {
				return invalid()
			}
		}
		switch q.Get("algorithm") {
		case "", "SHA1", "SHA256", "SHA512":
		default:
			return invalid()
		}
		if digits := q.Get("digits"); digits != "" && digits != "6" && digits != "8" {
			return invalid()
		}
		if period := q.Get("period"); period != "" {
			n, err := strconv.Atoi(period)
			if err != nil || n < 1 || n > 300 {
				return invalid()
			}
		}
		secret = q.Get("secret")
	}
	secret = strings.ToUpper(strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '-' {
			return -1
		}
		return r
	}, secret))
	secret = strings.TrimRight(secret, "=")
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil || len(decoded) == 0 || base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(decoded) != secret {
		return invalid()
	}
	if uri != nil {
		q := uri.Query()
		q.Set("secret", secret)
		uri.RawQuery = q.Encode()
		return uri.String(), nil
	}
	return secret, nil
}
