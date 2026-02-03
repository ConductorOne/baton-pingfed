package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-pingfed/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	roleAssignmentEntitlementName = "assigned"
)

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PingFederateClient
}

// roleResource convert a PingFederateRole into a Resource.
func roleResource(ctx context.Context, role *client.PingFederateRole) (*v2.Resource, error) {
	newRoleResource, err := resource.NewRoleResource(
		role.Name,
		resourceTypeRole,
		role.ID,
		[]resource.RoleTraitOption{},
	)
	if err != nil {
		return nil, err
	}

	return newRoleResource, nil
}

func (o *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeRole
}

func (o *roleBuilder) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	opts resource.SyncOpAttrs,
) (
	[]*v2.Resource,
	*resource.SyncOpResults,
	error,
) {
	roles, err := o.client.GetRoles(
		ctx,
	)
	if err != nil {
		return nil, nil, err
	}

	rv := make([]*v2.Resource, 0)
	for _, role := range roles {
		newResource, err := roleResource(ctx, &role)
		if err != nil {
			return nil, nil, err
		}

		rv = append(rv, newResource)
	}
	return rv, &resource.SyncOpResults{NextPageToken: "", Annotations: nil}, nil
}

func (o *roleBuilder) Entitlements(
	ctx context.Context,
	r *v2.Resource,
	_ resource.SyncOpAttrs,
) (
	[]*v2.Entitlement,
	*resource.SyncOpResults,
	error,
) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"Roles.Entitlements",
		zap.String("resource.DisplayName", r.DisplayName),
		zap.String("resource.Id.Resource", r.Id.Resource),
	)
	entitlements := []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			r,
			roleAssignmentEntitlementName,
			entitlement.WithGrantableTo(resourceTypeUser),
			entitlement.WithDisplayName(
				fmt.Sprintf("%s User Role", r.DisplayName),
			),
			entitlement.WithDescription(
				fmt.Sprintf("Has the %s role in PingFederate", r.DisplayName),
			),
		),
	}

	return entitlements, &resource.SyncOpResults{NextPageToken: "", Annotations: nil}, nil
}

type UserRoleGrant struct {
	UserID     string
	UserRoleID string
}

func (o *roleBuilder) Grants(
	ctx context.Context,
	r *v2.Resource,
	opts resource.SyncOpAttrs,
) (
	[]*v2.Grant,
	*resource.SyncOpResults,
	error,
) {
	assignments, err := o.client.GetRoleAssignments(
		ctx,
		r.Id.Resource,
	)
	if err != nil {
		return nil, nil, err
	}

	grants := make([]*v2.Grant, 0)
	for _, assignment := range assignments {
		grants = append(grants, grant.NewGrant(
			r,
			roleAssignmentEntitlementName,
			&v2.ResourceId{
				ResourceType: resourceTypeUser.Id,
				Resource:     assignment.Username,
			},
		))
	}
	return grants, &resource.SyncOpResults{NextPageToken: "", Annotations: nil}, nil
}

func (o *roleBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	if principal.Id.ResourceType != resourceTypeUser.Id {
		logger.Warn(
			"pingfederate-connector: only users can be granted roles",
			zap.String("principal_type", principal.Id.ResourceType),
			zap.String("principal_id", principal.Id.Resource),
		)
		return nil, fmt.Errorf("pingfederate-connector: only users can be granted roles")
	}

	err := o.client.AddUserToRole(
		ctx,
		principal.Id.Resource,
		entitlement.Resource.Id.Resource,
	)
	return nil, err
}

func (o *roleBuilder) Revoke(
	ctx context.Context,
	grant *v2.Grant,
) (annotations.Annotations, error) {
	err := o.client.RemoveUserFromRole(
		ctx,
		grant.Principal.Id.Resource,
		grant.Entitlement.Resource.Id.Resource,
	)
	return nil, err
}

func newRoleBuilder(client *client.PingFederateClient) *roleBuilder {
	return &roleBuilder{
		resourceType: resourceTypeRole,
		client:       client,
	}
}
