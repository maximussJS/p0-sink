package validate

import (
	"errors"
	"regexp"
	"strings"
)

type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

var blacklistedHeaders = map[string]struct{}{
	"host": {}, "connection": {}, "content-length": {}, "transfer-encoding": {},
	"content-encoding": {}, "content-type": {}, "expect": {}, "upgrade": {},
	"origin": {}, "cookie": {}, "set-cookie": {}, "proxy-authorization": {},
	"proxy-authenticate": {}, "te": {}, "cache-control": {}, "if-modified-since": {},
	"if-unmodified-since": {}, "if-match": {}, "if-none-match": {}, "x-forwarded-for": {},
	"x-forwarded-host": {}, "x-forwarded-proto": {}, "www-authenticate": {}, "user-agent": {},
	"vary": {}, "x-block-stream-id": {}, "x-block-stream-direction": {}, "x-block-stream-from-block": {},
	"x-block-stream-to-block": {}, "sec-fetch-site": {}, "sec-fetch-mode": {}, "sec-fetch-user": {},
	"sec-fetch-dest": {}, "strict-transport-security": {}, "content-security-policy": {},
	"x-content-security-policy": {}, "x-webkit-csp": {}, "x-frame-options": {}, "x-powered-by": {},
	"server": {}, "access-control-allow-origin": {}, "access-control-allow-credentials": {},
	"access-control-allow-headers": {}, "access-control-allow-methods": {}, "x-xss-protection": {},
	"referer": {}, "referrer-policy": {}, "from": {}, "via": {}, "warning": {}, "x-request-id": {},
	"x-correlation-id": {}, "forwarded": {},
}

func ValidateHeaders(headers map[string]string) ValidationResult {
	if err := validateHeadersObject(headers); err != nil {
		return ValidationResult{Valid: false, Message: err.Error()}
	}

	for key, value := range headers {
		if err := validateHeaderKey(key); err != nil {
			return ValidationResult{Valid: false, Message: err.Error()}
		}
		if err := validateHeaderValue(value); err != nil {
			return ValidationResult{Valid: false, Message: err.Error()}
		}
	}

	return ValidationResult{Valid: true}
}

func validateHeadersObject(headers map[string]string) error {
	if headers == nil {
		return errors.New("Headers are not an object")
	}

	size := len(headers)
	if size > 10 {
		return errors.New("Too many headers")
	}
	if size == 0 {
		return errors.New("No headers")
	}
	return nil
}

func validateHeaderKey(key string) error {
	if len(key) == 0 || len(key) > 256 {
		return errors.New("Header name length is out of range. It should be between 1 and 256 characters")
	}

	match, _ := regexp.MatchString(`^[A-Za-z][A-Za-z0-9-_]*[A-Za-z0-9]$`, key)
	if !match {
		return errors.New("Header name does not match the required pattern. It should start with a letter, contain only alphanumeric characters, underscores, hyphens, and end with a letter or digit")
	}

	if _, exists := blacklistedHeaders[strings.ToLower(key)]; exists {
		return errors.New("Header is blacklisted")
	}
	return nil
}

func validateHeaderValue(value string) error {
	if len(value) == 0 || len(value) > 8192 {
		return errors.New("Header value length is out of range. It should be between 1 and 8192 characters")
	}

	if strings.TrimSpace(value) != value {
		return errors.New("Header value should not have leading or trailing whitespace")
	}

	reg := regexp.MustCompile(`^[^\x00-\x1F\x7F\u200B\u202D\u202E\u200E\u2060\u200F\u2061-\u2064\uFEFF\uFFF9-\uFFFC]+$`)
	if !reg.MatchString(value) {
		return errors.New("Header value contains forbidden characters")
	}
	return nil
}
