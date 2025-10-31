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

type groupBuilder struct {
	*dropbox.Client
}

const groupMembership = "member"
const groupOwner = "owner"
const limit = 100

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

func (o *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, attr resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	logger := ctxzap.Extract(ctx)
	token := attr.PageToken.Token
	logger.Debug("Starting Groups List", zap.String("token", token))
	outResources := []*v2.Resource{}

	var payload *dropbox.ListGroupsPayload
	var rateLimitData *v2.RateLimitDescription
	var err error

	if token == "" {
		payload, rateLimitData, err = o.ListGroups(ctx, limit)
	} else {
		payload, rateLimitData, err = o.ListGroupsContinue(ctx, token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		return nil, &resourceSdk.SyncOpResults{
			Annotations: outAnnotations,
		}, fmt.Errorf("error listing groups: %w", err)
	}

	for _, group := range payload.Groups {
		groupResource, err := groupResource(group, parentResourceID)
		if err != nil {
			return nil, &resourceSdk.SyncOpResults{
				Annotations: outAnnotations,
			}, err
		}
		outResources = append(outResources, groupResource)
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

// Entitlements always returns an empty slice for roles.
func (o *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
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
	}, nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, attr resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	var outGrants []*v2.Grant
	var payload *dropbox.ListGroupMembersPayload
	var err error
	var rateLimitData *v2.RateLimitDescription

	token := attr.PageToken.Token
	if token == "" {
		payload, rateLimitData, err = o.ListGroupMembers(ctx, resource.Id.Resource, 0)
	} else {
		payload, rateLimitData, err = o.ListGroupMembersContinue(ctx, token)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		return nil, &resourceSdk.SyncOpResults{
			Annotations: outAnnotations,
		}, fmt.Errorf("error listing group members: %w", err)
	}

	for _, user := range payload.Members {
		principalId, err := resourceSdk.NewResourceID(userResourceType, user.Profile.TeamMemberID)
		if err != nil {
			return nil, &resourceSdk.SyncOpResults{
				Annotations: outAnnotations,
			}, fmt.Errorf("error creating principal ID: %w", err)
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

	return outGrants, &resourceSdk.SyncOpResults{
		NextPageToken: cursor,
		Annotations:   outAnnotations,
	}, nil
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
	l := ctxzap.Extract(ctx)
	groupId := entitlement.Resource.Id.Resource
	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can be granted group membership")
	}

	teamMemberID := principal.Id.Resource

	var rateLimitData *v2.RateLimitDescription
	var err error
	switch entitlement.Slug {
	case groupMembership:
		rateLimitData, err = r.AddUserToGroup(ctx, groupId, teamMemberID, groupMembership)
	case groupOwner:
		rateLimitData, err = r.AddUserToGroup(ctx, groupId, teamMemberID, groupOwner)
	}

	var outputAnnotations annotations.Annotations
	outputAnnotations.WithRateLimiting(rateLimitData)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			l.Warn("baton-dropbox: group membership to grant already exists; treating as successful because the end state is achieved",
				zap.String("group_id", groupId),
				zap.String("team_member_id", teamMemberID))
			return annotations.New(&v2.GrantAlreadyExists{}), nil
		}
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to add user to group: %w", err)
	}

	return outputAnnotations, nil
}

func (r *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	principal := grant.Principal
	entitlement := grant.Entitlement
	groupId := entitlement.Resource.Id.Resource

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-dropbox: only users can have group membership revoked")
	}

	teamMemberID := principal.Id.Resource

	ratelimitData, err := r.RemoveUserFromGroup(ctx, groupId, teamMemberID)
	var outputAnnotations annotations.Annotations
	outputAnnotations.WithRateLimiting(ratelimitData)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists || status.Code(err) == codes.NotFound {
			l.Warn("baton-dropbox: group membership to revoke not found; treating as successful because the end state is achieved",
				zap.String("group_id", groupId),
				zap.String("team_member_id", teamMemberID))
			return annotations.New(&v2.GrantAlreadyRevoked{}), nil
		}
		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership from group: %w", err)
	}
	return outputAnnotations, nil
}
