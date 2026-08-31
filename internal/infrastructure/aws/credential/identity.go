package credential

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Identity struct {
	AccountID string
	ARN       string
	UserID    string
}

func VerifyIdentity(ctx context.Context, cfg aws.Config) (*Identity, error) {
	client := sts.NewFromConfig(cfg)
	output, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to verify AWS identity: %w", err)
	}

	return &Identity{
		AccountID: *output.Account,
		ARN:       *output.Arn,
		UserID:    *output.UserId,
	}, nil
}
