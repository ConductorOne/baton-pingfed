package connector

import (
	"context"
	"fmt"

	"github.com/ConductorOne/baton-pingfed/pkg/connector/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.PingFederateClient
}

// userResource convert a PingFederateUser into a Resource.
func userResource(
	user client.PingFederateUser,
) (*v2.Resource, error) {
	displayName := user.Username
	status := v2.UserTrait_Status_STATUS_DISABLED
	if user.IsActive {
		status = v2.UserTrait_Status_STATUS_ENABLED
	}

	profile := map[string]interface{}{
		"username":    user.Username,
		"phoneNumber": user.PhoneNumber,
		"email":       user.Email,
		"isAuditor":   user.IsAuditor,
		"department":  user.Department,
		"description": user.Description,
	}

	userTraitOptions := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(status),
		resource.WithUserLogin(user.Username),
	}

	if user.Email != "" {
		userTraitOptions = append(userTraitOptions, resource.WithEmail(user.Email, true))
	}

	newUserResource, err := resource.NewUserResource(
		displayName,
		resourceTypeUser,
		user.Username,
		userTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return newUserResource, nil
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return resourceTypeUser
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (b *userBuilder) List(
	ctx context.Context,
	resourceID *v2.ResourceId,
	token *pagination.Token,
) (
	[]*v2.Resource,
	string,
	annotations.Annotations,
	error,
) {
	users, err := b.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list users: %w", err)
	}

	rv := make([]*v2.Resource, 0)
	for _, user := range users {
		ur, err := userResource(user)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, ur)
	}

	return rv, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(
	_ context.Context,
	resource *v2.Resource,
	_ *pagination.Token,
) (
	[]*v2.Entitlement,
	string,
	annotations.Annotations,
	error,
) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(
	ctx context.Context,
	resource *v2.Resource,
	pToken *pagination.Token,
) (
	[]*v2.Grant,
	string,
	annotations.Annotations,
	error,
) {
	return nil, "", nil, nil
}

func newUserBuilder(
	client *client.PingFederateClient,
) *userBuilder {
	return &userBuilder{
		resourceType: resourceTypeUser,
		client:       client,
	}
}
