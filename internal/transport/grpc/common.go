package grpc

import (
	"fmt"
	"strings"
)

type ResourceID string

func (r ResourceID) Validate() error {
	if strings.TrimSpace(string(r)) == "" {
		return fmt.Errorf("resource id must not be empty")
	}
	return nil
}

type ResourceCategory string

const (
	CategoryCompute     ResourceCategory = "compute"
	CategoryNetwork     ResourceCategory = "network"
	CategoryDatabase    ResourceCategory = "database"
	CategoryStorage     ResourceCategory = "storage"
	CategorySecurity    ResourceCategory = "security"
	CategoryIntegration ResourceCategory = "integration"
	CategoryOther       ResourceCategory = "other"
)

type ResourceType string

type Tag map[string]string
