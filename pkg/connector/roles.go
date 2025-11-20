package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (o *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, attr resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	logger := ctxzap.Extract(ctx)
	token := attr.PageToken.Token
	logger.Debug("Starting Roles List", zap.String("token", token))
	outResources := []*v2.Resource{}

	var payload *dropbox.ListUsersPayload
	var rateLimitData *v2.RateLimitDescription
	var err error

	if token == "" {
		payload, rateLimitData, err = o.ListUsers(ctx, limit)
	} else {
		payload, rateLimitData, err = o.ListUsersContinue(ctx, token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		return nil, &resourceSdk.SyncOpResults{
			Annotations: outAnnotations,
		}, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range payload.Members {
		for _, role := range user.Roles {
			roleResource, err := roleResource(role, parentResourceID)
			if err != nil {
				return nil, &resourceSdk.SyncOpResults{
					Annotations: outAnnotations,
				}, err
			}
			outResources = append(outResources, roleResource)
		}
	}

	var cursor string
	if payload.HasMore {
		cursor = payload.Cursor
	}

	return outResources, &resourceSdk.SyncOpResults{
		NextPageToken: cursor,
		Annotations:   outAnnotations,
	}, nil
}

func (o *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			roleMembership,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s Role %s", resource.DisplayName, roleMembership)),
			entitlement.WithDescription(fmt.Sprintf("Member of %s Dropbox role", resource.DisplayName)),
		),
	}, nil, nil
}

func (o *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, attr resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	var outGrants []*v2.Grant

	var payload *dropbox.ListUsersPayload
	var rateLimitData *v2.RateLimitDescription
	var err error

	token := attr.PageToken.Token
	if token == "" {
		payload, rateLimitData, err = o.ListUsers(ctx, limit)
	} else {
		payload, rateLimitData, err = o.ListUsersContinue(ctx, token)
	}
	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		return nil, &resourceSdk.SyncOpResults{
			Annotations: outAnnotations,
		}, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range payload.Members {
		if !user.HasRole(resource.Id.Resource) {
			continue
		}
		principalId, err := resourceSdk.NewResourceID(userResourceType, user.Profile.TeamMemberID)
		if err != nil {
			return nil, &resourceSdk.SyncOpResults{
				Annotations: outAnnotations,
			}, fmt.Errorf("error creating principal ID: %w", err)
		}
		nextGrant := grant.NewGrant(
			resource,
			roleMembership,
			principalId,
		)
		outGrants = append(outGrants, nextGrant)
	}

	var cursor string
	if payload.HasMore {
		cursor = payload.Cursor
	}
	return outGrants, &resourceSdk.SyncOpResults{
		NextPageToken: cursor,
		Annotations:   outAnnotations,
	}, nil
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
	l := ctxzap.Extract(ctx)
	roleId := entitlement.Resource.Id.Resource
	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can be granted role membership")
	}

	teamMemberID := principal.Id.Resource

	rateLimitData, err := r.AddRoleToUser(ctx, roleId, teamMemberID)
	var outputAnnotations annotations.Annotations
	outputAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			l.Warn("baton-dropbox: role membership to grant already exists; treating as successful because the end state is achieved",
				zap.String("role_id", roleId),
				zap.String("team_member_id", teamMemberID))
			return annotations.New(&v2.GrantAlreadyExists{}), nil
		}
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to add user to role: %w", err)
	}

	return outputAnnotations, nil
}

func (r *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	principal := grant.Principal

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can have role membership revoked")
	}

	teamMemberID := principal.Id.Resource

	var outputAnnotations annotations.Annotations
	ratelimitData, err := r.ClearRoles(ctx, teamMemberID)
	outputAnnotations.WithRateLimiting(ratelimitData)

	if err != nil {
		if status.Code(err) == codes.AlreadyExists || status.Code(err) == codes.NotFound {
			l.Warn("baton-dropbox: role membership to revoke not found; treating as successful because the end state is achieved",
				zap.String("team_member_id", teamMemberID))
			return annotations.New(&v2.GrantAlreadyRevoked{}), nil
		}
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership from role: %w", err)
	}
	return outputAnnotations, nil
}
