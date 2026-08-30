package resource

import "errors"

type ResourceID string

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

type Resource struct {
	id         ResourceID
	arn        string
	accountID  string
	region     string
	typ        ResourceType
	category   ResourceCategory
	name       string
	metadata   map[string]string
	tags       Tag
}

type ResourceOption func(*Resource)

func WithARN(arn string) ResourceOption {
	return func(r *Resource) {
		r.arn = arn
	}
}

func WithAccountID(accountID string) ResourceOption {
	return func(r *Resource) {
		r.accountID = accountID
	}
}

func WithRegion(region string) ResourceOption {
	return func(r *Resource) {
		r.region = region
	}
}

func WithMetadata(metadata map[string]string) ResourceOption {
	return func(r *Resource) {
		r.metadata = metadata
	}
}

func WithTags(tags Tag) ResourceOption {
	return func(r *Resource) {
		r.tags = tags
	}
}

func NewResource(id ResourceID, typ ResourceType, category ResourceCategory, name string, opts ...ResourceOption) (*Resource, error) {
	if id == "" {
		return nil, errors.New("resource id must not be empty")
	}
	if name == "" {
		return nil, errors.New("resource name must not be empty")
	}

	r := &Resource{
		id:       id,
		typ:      typ,
		category: category,
		name:     name,
		metadata: make(map[string]string),
		tags:     make(Tag),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Resource) ID() ResourceID {
	return r.id
}

func (r *Resource) ARN() string {
	return r.arn
}

func (r *Resource) AccountID() string {
	return r.accountID
}

func (r *Resource) Region() string {
	return r.region
}

func (r *Resource) Type() ResourceType {
	return r.typ
}

func (r *Resource) Category() ResourceCategory {
	return r.category
}

func (r *Resource) Name() string {
	return r.name
}

func (r *Resource) Metadata() map[string]string {
	return r.metadata
}

func (r *Resource) Tags() Tag {
	return r.tags
}
