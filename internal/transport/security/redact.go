package security

import (
	"regexp"
	"strings"
)

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(aws_access_key_id|aws_secret_access_key|password|secret|token|api_key|apikey)\s*[:=]\s*[^\s&]+`),
		regexp.MustCompile(`(?i)(authorization)\s*[:=]\s*[^\n]+`),
		regexp.MustCompile(`(?i)(Bearer|Basic|ApiKey)\s+[A-Za-z0-9+/=]{16,}`),
		regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`),
		regexp.MustCompile(`(?i)(secret.{0,20}[0-9a-zA-Z/+]{40})`),
	}

	sensitiveHeaders = map[string]bool{
		"authorization": true,
		"x-api-key":     true,
		"cookie":        true,
		"set-cookie":    true,
	}
)

const redactedValue = "[REDACTED]"

func RedactString(input string) string {
	result := input
	for _, pattern := range secretPatterns {
		result = pattern.ReplaceAllString(result, redactedValue)
	}
	return result
}

func RedactMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		if sensitiveHeaders[strings.ToLower(k)] {
			result[k] = redactedValue
		} else {
			result[k] = RedactString(v)
		}
	}
	return result
}

func RedactHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	for k, v := range headers {
		if sensitiveHeaders[strings.ToLower(k)] {
			result[k] = []string{redactedValue}
		} else {
			result[k] = v
		}
	}
	return result
}

func IsSensitiveHeader(name string) bool {
	return sensitiveHeaders[strings.ToLower(name)]
}
