package resource_test

import (
	"testing"

	"github.com/elip/WeaveLens/internal/domain/resource"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		testName    string
		id      resource.ResourceID
		typ     resource.ResourceType
		category resource.ResourceCategory
		resourceName    string
		opts    []resource.ResourceOption
		wantErr bool
	}{
		{
			testName:     "valid resource",
			id:       "res-1",
			typ:      "EC2",
			category: resource.CategoryCompute,
			resourceName:     "MyInstance",
			wantErr:  false,
		},
		{
			testName:     "empty id",
			id:       "",
			typ:      "EC2",
			category: resource.CategoryCompute,
			resourceName:     "MyInstance",
			wantErr:  true,
		},
		{
			testName:     "empty name",
			id:       "res-1",
			typ:      "EC2",
			category: resource.CategoryCompute,
			resourceName:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := resource.NewResource(tt.id, tt.typ, tt.category, tt.resourceName, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID() != tt.id {
					t.Errorf("NewResource() ID = %v, want %v", got.ID(), tt.id)
				}
				if got.Type() != tt.typ {
					t.Errorf("NewResource() Type = %v, want %v", got.Type(), tt.typ)
				}
				if got.Category() != tt.category {
					t.Errorf("NewResource() Category = %v, want %v", got.Category(), tt.category)
				}
				if got.Name() != tt.resourceName {
					t.Errorf("NewResource() Name = %v, want %v", got.Name(), tt.resourceName)
				}
			}
		})
	}
}

func TestResourceOptions(t *testing.T) {
	res, err := resource.NewResource(
		"res-1",
		"EC2",
		resource.CategoryCompute,
		"MyInstance",
		resource.WithARN("arn:aws:ec2:us-east-1:123456789012:instance/i-123"),
		resource.WithAccountID("123456789012"),
		resource.WithRegion("us-east-1"),
	)
	if err != nil {
		t.Fatalf("NewResource() error = %v", err)
	}

	if res.ARN() != "arn:aws:ec2:us-east-1:123456789012:instance/i-123" {
		t.Errorf("ARN = %v, want arn:aws:ec2:us-east-1:123456789012:instance/i-123", res.ARN())
	}
	if res.AccountID() != "123456789012" {
		t.Errorf("AccountID = %v, want 123456789012", res.AccountID())
	}
	if res.Region() != "us-east-1" {
		t.Errorf("Region = %v, want us-east-1", res.Region())
	}
}
