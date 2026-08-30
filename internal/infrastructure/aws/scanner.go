package aws

import (
	"context"

	"github.com/elip/WeaveLens/internal/domain/resource"
)

type Scanner interface {
	Scan(ctx context.Context) ([]*resource.Resource, error)
}
