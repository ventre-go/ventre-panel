// Package secure provides password redaction and memory safety helpers.
package secure

import "strings"

// RedactSecrets replaces all occurrences of any secret string in s with "(redacted)".
// This must be called before displaying any error, result, or log that might contain
// a password.
func RedactSecrets(s string, secrets []string) string {
	result := s
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		result = strings.ReplaceAll(result, secret, "(redacted)")
	}
	return result
}

// MaskPassword returns a masked representation of a password.
func MaskPassword(p string) string {
	if p == "" {
		return ""
	}
	return strings.Repeat("•", len(p))
}

// SanitizeErrorForDisplay removes potential secrets from an error message.
func SanitizeErrorForDisplay(errMsg string, secrets []string) string {
	return RedactSecrets(errMsg, secrets)
}
