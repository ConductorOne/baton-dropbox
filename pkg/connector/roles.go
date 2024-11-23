package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type roleBuilder struct {
	*dropbox.Client
}

const roleMembership = "member"

func roleResource(role dropbox.Role, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	return resourceSdk.NewRoleResource(
		role.Name,
		roleResourceType,
		role.RoleID,
		[]resourceSdk.RoleTraitOption{
			resourceSdk.WithRoleProfile(
				map[string]interface{}{
					"id":          role.RoleID,
					"name":        role.Name,
					"description": role.Description,
				},
			),
		},
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

func (o *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

func (o *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug("Starting Roles List", zap.String("token", pToken.Token))
	outResources := []*v2.Resource{}

	users, err := o.ListUsers(ctx, 0, false)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range users.Members {
		for _, role := range user.Roles {
			roleResource, err := roleResource(role, parentResourceID)
			if err != nil {
				return nil, "", nil, err
			}
			outResources = append(outResources, roleResource)
		}
	}

	return outResources, "", nil, nil
}

func (o *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			roleMembership,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s Role %s", resource.DisplayName, roleMembership)),
			entitlement.WithDescription(fmt.Sprintf("Member of %s Dropbox role", resource.DisplayName)),
		),
	}, "", nil, nil
}

func (o *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var outGrants []*v2.Grant
	users, err := o.ListUsers(ctx, 0, false)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range users.Members {
		if !user.HasRole(resource.Id.Resource) {
			continue
		}
		principalId, err := resourceSdk.NewResourceID(userResourceType, user.Profile.AccountID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating principal ID: %w", err)
		}
		nextGrant := grant.NewGrant(
			resource,
			roleMembership,
			principalId,
		)
		outGrants = append(outGrants, nextGrant)
	}
	return outGrants, "", nil, nil
}

func newRoleBuilder(client *dropbox.Client) *roleBuilder {
	return &roleBuilder{
		Client: client,
	}
}

func (r *roleBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (
	annotations.Annotations,
	error,
) {
	userId := principal.Id.Resource
	roleId := entitlement.Resource.Id.Resource
	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can be granted role membership")
	}

	err := r.AddRoleToUser(ctx, roleId, userId)
	if err != nil {
		return nil, fmt.Errorf("baton-dropbox: failed to add user to role: %s", err.Error())
	}

	return nil, nil
}

func (r *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	principal := grant.Principal
	userId := principal.Id.Resource

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-auth0: only users can have role membership revoked")
	}

	var outputAnnotations annotations.Annotations
	ratelimitData, err := r.ClearRoles(ctx, userId)
	outputAnnotations.WithRateLimiting(ratelimitData)

	if err != nil {
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership to role: %s", err.Error())
	}
	return outputAnnotations, nil
}
