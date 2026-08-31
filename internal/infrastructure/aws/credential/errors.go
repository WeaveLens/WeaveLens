package credential

import (
	"errors"

	"github.com/aws/smithy-go"
)

var (
	ErrMissingCredentials = errors.New("missing AWS credentials")
	ErrInvalidCredentials = errors.New("invalid AWS credentials")
	ErrExpiredCredentials = errors.New("expired AWS credentials")
	ErrAssumeRole         = errors.New("failed to assume role")
	ErrAccessDenied       = errors.New("AWS access denied")
	ErrInvalidRoleARN     = errors.New("invalid Role ARN")
	ErrSTSFailure         = errors.New("STS call failed")
)

func isAccessDenied(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "AccessDenied"
	}
	return false
}
