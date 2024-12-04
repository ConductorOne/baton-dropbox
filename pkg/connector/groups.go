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
const groupOwner = "owner"

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

	var payload *dropbox.ListGroupsPayload
	var rateLimitData *v2.RateLimitDescription
	var err error
	var limit int = 100

	if pToken.Token == "" {
		payload, rateLimitData, err = o.ListGroups(ctx, limit)
	} else {
		payload, rateLimitData, err = o.ListGroupsContinue(ctx, pToken.Token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		return nil, "", outAnnotations, fmt.Errorf("error listing groups: %w", err)
	}

	for _, group := range payload.Groups {
		groupResource, err := groupResource(group, parentResourceID)
		if err != nil {
			return nil, "", outAnnotations, err
		}
		outResources = append(outResources, groupResource)
	}

	var cursor string
	if payload.HasMore {
		cursor = payload.Cursor
	}

	return outResources, cursor, outAnnotations, nil
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
		entitlement.NewAssignmentEntitlement(
			resource,
			groupOwner,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(fmt.Sprintf("Owner of %s dropbox group", resource.DisplayName)),
			entitlement.WithDisplayName(fmt.Sprintf("%s group %s", resource.DisplayName, groupOwner)),
		),
	}, "", nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var outGrants []*v2.Grant
	var payload *dropbox.ListGroupMembersPayload
	var err error
	var rateLimitData *v2.RateLimitDescription

	if pToken.Token == "" {
		payload, rateLimitData, err = o.ListGroupMembers(ctx, resource.Id.Resource, 0)
	} else {
		payload, rateLimitData, err = o.ListGroupMembersContinue(ctx, pToken.Token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		return nil, "", outAnnotations, fmt.Errorf("error listing group members: %w", err)
	}

	for _, user := range payload.Members {
		principalId, err := resourceSdk.NewResourceID(userResourceType, user.Profile.AccountID)
		if err != nil {
			return nil, "", outAnnotations, fmt.Errorf("error creating principal ID: %w", err)
		}

		var nextGrant *v2.Grant
		switch user.AccessType.Tag {
		case groupMembership:
			nextGrant = grant.NewGrant(
				resource,
				groupMembership,
				principalId,
			)
		case groupOwner:
			nextGrant = grant.NewGrant(
				resource,
				groupOwner,
				principalId,
			)
		}
		outGrants = append(outGrants, nextGrant)
	}

	var cursor string
	if payload.HasMore {
		cursor = payload.Cursor
	}

	return outGrants, cursor, outAnnotations, nil
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
	groupId := entitlement.Resource.Id.Resource
	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can be granted role membership")
	}

	email, err := getEmail(principal)
	if err != nil {
		return nil, fmt.Errorf("baton-auth0: failed to get email for user: %w", err)
	}

	var rateLimitData *v2.RateLimitDescription
	switch entitlement.Slug {
	case groupMembership:
		rateLimitData, err = r.AddUserToGroup(ctx, groupId, email, groupMembership)
	case groupOwner:
		rateLimitData, err = r.AddUserToGroup(ctx, groupId, email, groupOwner)
	}

	var outputAnnotations annotations.Annotations
	outputAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to add user to role: %w", err)
	}

	return outputAnnotations, nil
}

func (r *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	principal := grant.Principal
	entitlement := grant.Entitlement
	groupId := entitlement.Resource.Id.Resource

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-auth0: only users can have role membership revoked")
	}

	email, err := getEmail(principal)
	if err != nil {
		return nil, fmt.Errorf("baton-auth0: failed to get email for user: %w", err)
	}

	ratelimitData, err := r.RemoveUserFromGroup(ctx, groupId, email)
	var outputAnnotations annotations.Annotations
	outputAnnotations.WithRateLimiting(ratelimitData)
	if err != nil {
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership to group: %w", err)
	}
	return outputAnnotations, nil
}
