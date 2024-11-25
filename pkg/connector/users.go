package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type userBuilder struct {
	*dropbox.Client
}

func userResource(user dropbox.Profile, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":         user.AccountID,
		"email":      user.Email,
		"first_name": user.Name.GivenName,
		"last_name":  user.Name.Surname,
	}

	userTraitOptions := []resourceSdk.UserTraitOption{
		resourceSdk.WithEmail(user.Email, true),
		resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
		resourceSdk.WithUserProfile(profile),
		resourceSdk.WithUserLogin(user.Email),
	}

	return resourceSdk.NewUserResource(
		user.Email,
		userResourceType,
		user.AccountID,
		userTraitOptions,
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug("Starting Users List", zap.String("token", pToken.Token))

	outResources := []*v2.Resource{}
	var payload *dropbox.ListUsersPayload
	var rateLimitData *v2.RateLimitDescription
	var err error
	var limit int = 100

	if pToken == nil {
		payload, rateLimitData, err = o.ListUsers(ctx, limit)
	} else {
		payload, rateLimitData, err = o.ListUsersContinue(ctx, pToken.Token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		return nil, "", outAnnotations, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range payload.Members {
		resource, err := userResource(user.Profile, parentResourceID)
		if err != nil {
			return nil, "", outAnnotations, err
		}
		outResources = append(outResources, resource)
	}

	var cursor string
	if payload.HasMore {
		cursor = payload.Cursor
	}

	return outResources, cursor, outAnnotations, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *dropbox.Client) *userBuilder {
	return &userBuilder{
		Client: client,
	}
}
