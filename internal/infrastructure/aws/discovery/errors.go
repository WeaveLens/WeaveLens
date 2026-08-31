package discovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/smithy-go"
)

var (
	ErrAccessDenied     = errors.New("AWS access denied")
	ErrThrottling       = errors.New("AWS API throttling")
	ErrTransientFailure = errors.New("transient AWS failure")
	ErrPartialFailure   = errors.New("partial scan failure")
	ErrContextCanceled  = errors.New("scan context canceled")
)

type ScannerError struct {
	Scanner string
	Err     error
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("scanner %s failed: %v", e.Scanner, e.Err)
}

func (e *ScannerError) Unwrap() error {
	return e.Err
}

func ClassifyError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "AccessDenied", "UnauthorizedOperation":
			return fmt.Errorf("%w: %v", ErrAccessDenied, err)
		case "Throttling", "RequestLimitExceeded", "ThrottlingException":
			return fmt.Errorf("%w: %v", ErrThrottling, err)
		case "ServiceUnavailable", "InternalFailure", "InternalServerError":
			return fmt.Errorf("%w: %v", ErrTransientFailure, err)
		}
	}

	if errors.Is(err, context.Canceled) {
		return ErrContextCanceled
	}

	return err
}
