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

type groupBuilder struct {
	*dropbox.Client
}

const groupMembership = "member"

func groupResource(group dropbox.Group, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	return resourceSdk.NewGroupResource(
		group.Name,
		groupResourceType,
		group.GroupID,
		[]resourceSdk.GroupTraitOption{
			resourceSdk.WithGroupProfile(
				map[string]interface{}{
					"id":   group.GroupID,
					"name": group.Name,
				},
			),
		},
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

func (o *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (o *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug("Starting Groups List", zap.String("token", pToken.Token))
	outResources := []*v2.Resource{}

	groups, err := o.ListGroups(ctx, 0)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error listing users: %w", err)
	}

	for _, group := range groups.Groups {
		groupResource, err := groupResource(group, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}
		outResources = append(outResources, groupResource)
	}

	return outResources, "", nil, nil
}

// Entitlements always returns an empty slice for roles.
func (o *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			resource,
			groupMembership,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(fmt.Sprintf("Member of %s Dropbox group", resource.DisplayName)),
			entitlement.WithDisplayName(fmt.Sprintf("%s Group %s", resource.DisplayName, groupMembership)),
		),
	}, "", nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var outGrants []*v2.Grant
	group, err := o.ListGroupMembers(ctx, resource.Id.Resource, 0)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range group.Members {
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

func newGroupBuilder(client *dropbox.Client) *groupBuilder {
	return &groupBuilder{
		Client: client,
	}
}

func (r *groupBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (
	annotations.Annotations,
	error,
) {
	// userId := principal.Id.Resource
	// roleId := entitlement.Resource.Id.Resource
	// if principal.Id.ResourceType != userResourceType.Id {
	// 	return nil, fmt.Errorf("baton-dropbox: only users can be granted role membership")
	// }
	//
	// err := r.AddRoleToUser(ctx, roleId, userId)
	// if err != nil {
	// 	return nil, fmt.Errorf("baton-dropbox: failed to add user to role: %s", err.Error())
	// }

	return nil, nil
}

func (r *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	// principal := grant.Principal
	// userId := principal.Id.Resource
	//
	// if principal.Id.ResourceType != userResourceType.Id {
	// 	return nil, fmt.Errorf("baton-auth0: only users can have role membership revoked")
	// }
	//
	// var outputAnnotations annotations.Annotations
	// ratelimitData, err := r.ClearRoles(ctx, userId)
	// outputAnnotations.WithRateLimiting(ratelimitData)
	//
	// if err != nil {
	// 	return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership to role: %s", err.Error())
	// }
	// return outputAnnotations, nil
	return nil, nil
}
