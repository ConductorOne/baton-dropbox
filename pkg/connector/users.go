package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type userBuilder struct {
	*dropbox.Client
}

// mapUserStatus converts Dropbox user status to SDK status.
func mapUserStatus(status dropbox.Tag) v2.UserTrait_Status_Status {
	switch status.Tag {
	case "active":
		return v2.UserTrait_Status_STATUS_ENABLED
	case "invited":
		return v2.UserTrait_Status_STATUS_ENABLED
	case "suspended":
		return v2.UserTrait_Status_STATUS_DISABLED
	case "removed":
		return v2.UserTrait_Status_STATUS_DELETED
	default:
		return v2.UserTrait_Status_STATUS_UNSPECIFIED
	}
}

func userResource(user dropbox.Profile, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":             user.AccountID,
		"email":          user.Email,
		"first_name":     user.Name.GivenName,
		"last_name":      user.Name.Surname,
		"team_member_id": user.TeamMemberID,
		"status":         user.Status.Tag,
	}

	userStatus := mapUserStatus(user.Status)

	userTraitOptions := []resourceSdk.UserTraitOption{
		resourceSdk.WithEmail(user.Email, true),
		resourceSdk.WithStatus(userStatus),
		resourceSdk.WithUserProfile(profile),
		resourceSdk.WithUserLogin(user.Email),
	}

	return resourceSdk.NewUserResource(
		user.Email,
		userResourceType,
		user.TeamMemberID,
		userTraitOptions,
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, attr resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	logger := ctxzap.Extract(ctx)
	token := attr.PageToken.Token
	logger.Debug("Starting Users List", zap.String("token", attr.PageToken.Token))

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
		resource, err := userResource(user.Profile, parentResourceID)
		if err != nil {
			return nil, &resourceSdk.SyncOpResults{
				Annotations: outAnnotations,
			}, err
		}
		outResources = append(outResources, resource)
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

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newUserBuilder(client *dropbox.Client) *userBuilder {
	return &userBuilder{
		Client: client,
	}
}

// CreateAccountCapabilityDetails declares support for account provisioning without passwords.
func (o *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

// CreateAccount provisions a new user in Dropbox Team based on AccountInfo.
func (o *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	credentialOptions *v2.LocalCredentialOptions,
) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	l := ctxzap.Extract(ctx)
	profile := accountInfo.GetProfile().AsMap()

	email, ok := profile["email"].(string)
	if !ok || email == "" {
		return nil, nil, nil, fmt.Errorf("email is required")
	}

	response, rateLimitData, err := o.AddMember(ctx, email)
	var annos annotations.Annotations
	annos.WithRateLimiting(rateLimitData)

	if err != nil {
		l.Error("error creating user", zap.Error(err))
		return nil, nil, annos, err
	}

	if len(response.Complete) == 0 || response.Complete[0].Tag != "success" {
		return nil, nil, annos, fmt.Errorf("failed to create user: unexpected response")
	}

	newUserProfile := response.Complete[0].Profile
	newUserResource, err := userResource(newUserProfile, nil)
	if err != nil {
		l.Error("error converting created user to resource", zap.Error(err))
		return nil, nil, annos, err
	}

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: newUserResource,
	}, []*v2.PlaintextData{}, annos, nil
}

// Delete implements account deprovisioning for users.
func (o *userBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if resourceId.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("invalid resource type: expected %s, got %s", userResourceType.Id, resourceId.ResourceType)
	}

	teamMemberID := resourceId.Resource

	_, rateLimitData, err := o.RemoveMember(ctx, teamMemberID)
	var annos annotations.Annotations
	annos.WithRateLimiting(rateLimitData)

	if err != nil {
		l.Error("error deleting user", zap.Error(err), zap.String("teamMemberID", teamMemberID))
		return annos, err
	}

	return annos, nil
}
