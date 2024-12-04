package connector

//
// import (
// 	"context"
// 	"fmt"
//
// 	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
// 	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
// 	"github.com/conductorone/baton-sdk/pkg/annotations"
// 	"github.com/conductorone/baton-sdk/pkg/pagination"
// 	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
// 	"github.com/conductorone/baton-sdk/pkg/types/grant"
// 	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
// 	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
// 	"go.uber.org/zap"
// )
//
// type folderBuilder struct {
// 	*dropbox.Client
// }
//
// const folderOwner = "owner"
// const folderEditor = "editor"
// const folderViewer = "viewer"
// const folderViewerNoComment = "viewer_no_comment"
// const folderTraverse = "traverse"
// const folderNoAccess = "no_access"
//
// func folderResource(folder dropbox.Folder, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
// 	return resourceSdk.NewGroupResource(
// 		folder.Name,
// 		folderResourceType,
// 		folder.SharedFolderId,
// 		[]resourceSdk.GroupTraitOption{
// 			resourceSdk.WithGroupProfile(
// 				map[string]interface{}{
// 					"id":   folder.SharedFolderId,
// 					"name": folder.Name,
// 				},
// 			),
// 		},
// 		resourceSdk.WithParentResourceID(parentResourceID),
// 	)
// }
//
// func (o *folderBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
// 	return folderResourceType
// }
//
// func (o *folderBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
// 	logger := ctxzap.Extract(ctx)
// 	logger.Debug("Starting Folder List", zap.String("token", pToken.Token))
// 	outResources := []*v2.Resource{}
//
// 	var payload *dropbox.ListFoldersPayload
// 	var rateLimitData *v2.RateLimitDescription
// 	var err error
//
// 	if pToken.Token == "" {
// 		payload, rateLimitData, err = o.ListFolders(ctx)
// 	} else {
// 		payload, rateLimitData, err = o.ListFoldersContinue(ctx, pToken.Token)
// 	}
//
// 	var outAnnotations annotations.Annotations
// 	outAnnotations.WithRateLimiting(rateLimitData)
// 	if err != nil {
// 		return nil, "", outAnnotations, fmt.Errorf("error listing groups: %w", err)
// 	}
//
// 	for _, folder := range payload.Entries {
// 		folderResource, err := folderResource(folder, parentResourceID)
// 		if err != nil {
// 			return nil, "", outAnnotations, err
// 		}
// 		outResources = append(outResources, folderResource)
// 	}
//
// 	// endpoint keeps providing cursors infinitely, and there's not HasMore field being returned
// 	var cursor string
// 	if len(payload.Entries) != 0 {
// 		cursor = payload.Cursor
// 	}
// 	return outResources, cursor, outAnnotations, nil
// }
//
// // Entitlements always returns an empty slice for roles.
// func (o *folderBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
// 	return []*v2.Entitlement{
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderOwner,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Owner of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderOwner)),
// 		),
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderEditor,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Editor of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderEditor)),
// 		),
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderViewer,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Editor of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderViewer)),
// 		),
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderViewerNoComment,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Editor of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderViewerNoComment)),
// 		),
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderTraverse,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Editor of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderTraverse)),
// 		),
// 		entitlement.NewAssignmentEntitlement(
// 			resource,
// 			folderNoAccess,
// 			entitlement.WithGrantableTo(userResourceType, groupResourceType),
// 			entitlement.WithDescription(fmt.Sprintf("Editor of %s folder", resource.DisplayName)),
// 			entitlement.WithDisplayName(fmt.Sprintf("%s folder %s", resource.DisplayName, folderNoAccess)),
// 		),
// 	}, "", nil, nil
// }
//
// func (o *folderBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
// 	return nil, "", nil, nil
// }
//
// func (o *folderBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
// 	var outGrants []*v2.Grant
// 	var payload *dropbox.ListFolderMembersPayload
// 	var err error
// 	var rateLimitData *v2.RateLimitDescription
//
// 	if pToken.Token == "" {
// 		payload, rateLimitData, err = o.ListFolderMembers(ctx, resource.Id.Resource, 0)
// 	} else {
// 		payload, rateLimitData, err = o.ListFolderMembersContinue(ctx, pToken.Token)
// 	}
//
// 	var outAnnotations annotations.Annotations
// 	outAnnotations.WithRateLimiting(rateLimitData)
// 	if err != nil {
// 		return nil, "", outAnnotations, fmt.Errorf("error listing group members: %w", err)
// 	}
//
// 	for _, user := range payload.Users {
// 		principalId, err := resourceSdk.NewResourceID(userResourceType, user.User.AccountID)
// 		if err != nil {
// 			return nil, "", outAnnotations, fmt.Errorf("error creating principal ID: %w", err)
// 		}
//
// 		var nextGrant *v2.Grant
// 		switch user.AccessType.Tag {
// 		case folderOwner:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderOwner,
// 				principalId,
// 			)
// 		case folderEditor:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderEditor,
// 				principalId,
// 			)
// 		case folderViewer:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderViewer,
// 				principalId,
// 			)
// 		case folderViewerNoComment:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderViewerNoComment,
// 				principalId,
// 			)
// 		case folderTraverse:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderTraverse,
// 				principalId,
// 			)
// 		case folderNoAccess:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderNoAccess,
// 				principalId,
// 			)
// 		}
// 		outGrants = append(outGrants, nextGrant)
// 	}
//
// 	for _, group := range payload.Groups {
// 		principalId, err := resourceSdk.NewResourceID(userResourceType, group.Group.GroupID)
// 		if err != nil {
// 			return nil, "", outAnnotations, fmt.Errorf("error creating principal ID: %w", err)
// 		}
//
// 		var nextGrant *v2.Grant
// 		switch group.AccessType.Tag {
// 		case folderOwner:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderOwner,
// 				principalId,
// 			)
// 		case folderEditor:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderEditor,
// 				principalId,
// 			)
// 		case folderViewer:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderViewer,
// 				principalId,
// 			)
// 		case folderViewerNoComment:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderViewerNoComment,
// 				principalId,
// 			)
// 		case folderTraverse:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderTraverse,
// 				principalId,
// 			)
// 		case folderNoAccess:
// 			nextGrant = grant.NewGrant(
// 				resource,
// 				folderNoAccess,
// 				principalId,
// 			)
// 		}
// 		outGrants = append(outGrants, nextGrant)
// 	}
//
// 	return outGrants, payload.Cursor, outAnnotations, nil
// }
//
// func newFolderBuilder(client *dropbox.Client) *folderBuilder {
// 	return &folderBuilder{
// 		Client: client,
// 	}
// }
//
// func (r *folderBuilder) Grant(
// 	ctx context.Context,
// 	principal *v2.Resource,
// 	entitlement *v2.Entitlement,
// ) (
// 	annotations.Annotations,
// 	error,
// ) {
// 	groupId := entitlement.Resource.Id.Resource
// 	if principal.Id.ResourceType != userResourceType.Id {
// 		return nil, fmt.Errorf("baton-dropbox: only users can be granted role membership")
// 	}
//
// 	email, err := getEmail(principal)
// 	if err != nil {
// 		return nil, fmt.Errorf("baton-auth0: failed to get email for user: %w", err)
// 	}
//
// 	var rateLimitData *v2.RateLimitDescription
// 	switch entitlement.Slug {
// 	case groupMembership:
// 		rateLimitData, err = r.AddUserToGroup(ctx, groupId, email, groupMembership)
// 	case groupOwner:
// 		rateLimitData, err = r.AddUserToGroup(ctx, groupId, email, groupOwner)
// 	}
//
// 	var outputAnnotations annotations.Annotations
// 	outputAnnotations.WithRateLimiting(rateLimitData)
// 	if err != nil {
// 		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to add user to role: %w", err)
// 	}
//
// 	return outputAnnotations, nil
// }
//
// func (r *folderBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
// 	principal := grant.Principal
// 	entitlement := grant.Entitlement
// 	groupId := entitlement.Resource.Id.Resource
//
// 	if principal.Id.ResourceType != userResourceType.Id {
// 		return nil, fmt.Errorf("baton-auth0: only users can have role membership revoked")
// 	}
//
// 	email, err := getEmail(principal)
// 	if err != nil {
// 		return nil, fmt.Errorf("baton-auth0: failed to get email for user: %w", err)
// 	}
//
// 	ratelimitData, err := r.RemoveUserFromGroup(ctx, groupId, email)
// 	var outputAnnotations annotations.Annotations
// 	outputAnnotations.WithRateLimiting(ratelimitData)
// 	if err != nil {
// 		return outputAnnotations, fmt.Errorf("baton-dropbox: failed to revoke membership to group: %w", err)
// 	}
// 	return outputAnnotations, nil
// }
