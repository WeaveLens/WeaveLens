package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elip/WeaveLens/internal/domain/resource"
)

type IAMScanner struct {
	client IAMAPI
	region string
}

func NewIAMScanner(client IAMAPI, region string) *IAMScanner {
	return &IAMScanner{client: client, region: region}
}

func (s *IAMScanner) Name() string {
	return "IAM"
}

func (s *IAMScanner) Scan(ctx context.Context) ([]*resource.Resource, error) {
	var resources []*resource.Resource

	users, err := s.scanUsers(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "IAMUser", Err: err}
	}
	resources = append(resources, users...)

	roles, err := s.scanRoles(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "IAMRole", Err: err}
	}
	resources = append(resources, roles...)

	groups, err := s.scanGroups(ctx)
	if err != nil {
		return nil, &ScannerError{Scanner: "IAMGroup", Err: err}
	}
	resources = append(resources, groups...)

	return resources, nil
}

func (s *IAMScanner) scanUsers(ctx context.Context) ([]*resource.Resource, error) {
	paginator := iam.NewListUsersPaginator(s.client, &iam.ListUsersInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, user := range page.Users {
			if user.UserName == nil {
				continue
			}

			name := safePtr(user.UserName)
			metadata := map[string]string{}
			if user.Arn != nil {
				metadata["arn"] = *user.Arn
			}
			if user.UserId != nil {
				metadata["user_id"] = *user.UserId
			}
			if user.CreateDate != nil {
				metadata["create_date"] = user.CreateDate.Format("2006-01-02T15:04:05Z")
			}
			if user.Path != nil {
				metadata["path"] = *user.Path
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("IAMUser"),
				resource.CategorySecurity,
				name,
				resource.WithMetadata(metadata),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *IAMScanner) scanRoles(ctx context.Context) ([]*resource.Resource, error) {
	paginator := iam.NewListRolesPaginator(s.client, &iam.ListRolesInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, role := range page.Roles {
			if role.RoleName == nil {
				continue
			}

			name := safePtr(role.RoleName)
			metadata := map[string]string{}
			if role.Arn != nil {
				metadata["arn"] = *role.Arn
			}
			if role.RoleId != nil {
				metadata["role_id"] = *role.RoleId
			}
			if role.AssumeRolePolicyDocument != nil {
				metadata["assume_role_policy"] = *role.AssumeRolePolicyDocument
			}
			if role.CreateDate != nil {
				metadata["create_date"] = role.CreateDate.Format("2006-01-02T15:04:05Z")
			}
			if role.Path != nil {
				metadata["path"] = *role.Path
			}
			if role.MaxSessionDuration != nil {
				metadata["max_session_duration"] = fmt.Sprintf("%d", *role.MaxSessionDuration)
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("IAMRole"),
				resource.CategorySecurity,
				name,
				resource.WithMetadata(metadata),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

func (s *IAMScanner) scanGroups(ctx context.Context) ([]*resource.Resource, error) {
	paginator := iam.NewListGroupsPaginator(s.client, &iam.ListGroupsInput{})
	var resources []*resource.Resource

	for paginator.HasMorePages() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrContextCanceled, ctx.Err())
		default:
		}

		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, ClassifyError(err)
		}

		for _, group := range page.Groups {
			if group.GroupName == nil {
				continue
			}

			name := safePtr(group.GroupName)
			metadata := map[string]string{}
			if group.Arn != nil {
				metadata["arn"] = *group.Arn
			}
			if group.GroupId != nil {
				metadata["group_id"] = *group.GroupId
			}
			if group.CreateDate != nil {
				metadata["create_date"] = group.CreateDate.Format("2006-01-02T15:04:05Z")
			}
			if group.Path != nil {
				metadata["path"] = *group.Path
			}

			res, err := resource.NewResource(
				resource.ResourceID(name),
				resource.ResourceType("IAMGroup"),
				resource.CategorySecurity,
				name,
				resource.WithMetadata(metadata),
				resource.WithRegion(s.region),
			)
			if err != nil {
				continue
			}
			resources = append(resources, res)
		}
	}
	return resources, nil
}

type IAMAPI interface {
	ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error)
	ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error)
	ListGroups(ctx context.Context, params *iam.ListGroupsInput, optFns ...func(*iam.Options)) (*iam.ListGroupsOutput, error)
}
