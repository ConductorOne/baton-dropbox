package connector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const usersCacheTTL = 5 * time.Minute

type userBuilder struct {
	*dropbox.Client
	users          map[string]string
	usersMutex     sync.RWMutex
	usersTimestamp time.Time
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
		user.AccountID,
		userTraitOptions,
		resourceSdk.WithParentResourceID(parentResourceID),
	)
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// refreshUserCache updates the local user cache if TTL expired by fetching users from Dropbox.
func (o *userBuilder) refreshUserCache(ctx context.Context) error {
	o.usersMutex.Lock()
	defer o.usersMutex.Unlock()

	if o.users != nil && time.Since(o.usersTimestamp) < usersCacheTTL {
		return nil
	}

	o.users = make(map[string]string)
	var limit = 100

	payload, _, err := o.ListUsers(ctx, limit)
	if err != nil {
		return fmt.Errorf("dropbox-connector: failed to load users for cache: %w", err)
	}

	for _, member := range payload.Members {
		o.users[member.Profile.AccountID] = member.Profile.Email
	}

	for payload.HasMore {
		payload, _, err = o.ListUsersContinue(ctx, payload.Cursor)
		if err != nil {
			return fmt.Errorf("dropbox-connector: failed to load users for cache: %w", err)
		}

		for _, member := range payload.Members {
			o.users[member.Profile.AccountID] = member.Profile.Email
		}
	}

	o.usersTimestamp = time.Now()
	return nil
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)
	logger.Debug("Starting Users List", zap.String("token", pToken.Token))

	if err := o.refreshUserCache(ctx); err != nil {
		logger.Warn("Failed to refresh user cache", zap.Error(err))
	}

	outResources := []*v2.Resource{}
	var payload *dropbox.ListUsersPayload
	var rateLimitData *v2.RateLimitDescription
	var err error
	var limit = 100

	if pToken.Token == "" {
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

func getEmail(principal *v2.Resource) (string, error) {
	userTrait, err := resourceSdk.GetUserTrait(principal)
	if err != nil {
		return "", err
	}

	for _, email := range userTrait.GetEmails() {
		if email.IsPrimary {
			return email.Address, nil
		}
	}
	return "", fmt.Errorf("no primary email found for user")
}

// getEmailByAccountID finds a user by account_id and returns their email from the cache.
func (o *userBuilder) getEmailByAccountID(ctx context.Context, accountID string) (string, error) {
	if err := o.refreshUserCache(ctx); err != nil {
		return "", fmt.Errorf("failed to refresh user cache: %w", err)
	}

	o.usersMutex.RLock()
	email, found := o.users[accountID]
	o.usersMutex.RUnlock()

	if !found {
		return "", fmt.Errorf("user with account_id %s not found in cache", accountID)
	}

	return email, nil
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
	userResource, err := userResource(newUserProfile, nil)
	if err != nil {
		l.Error("error converting created user to resource", zap.Error(err))
		return nil, nil, annos, err
	}

	o.usersMutex.Lock()
	if o.users != nil {
		o.users[newUserProfile.AccountID] = newUserProfile.Email
	}
	o.usersMutex.Unlock()

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: userResource,
	}, []*v2.PlaintextData{}, annos, nil
}

// Delete implements account deprovisioning for users.
func (o *userBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if resourceId.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("invalid resource type: expected %s, got %s", userResourceType.Id, resourceId.ResourceType)
	}

	accountID := resourceId.Resource

	userEmail, err := o.getEmailByAccountID(ctx, accountID)
	if err != nil {
		l.Error("error getting email for user", zap.Error(err), zap.String("accountID", accountID))
		return nil, fmt.Errorf("failed to get email for account_id %s: %w", accountID, err)
	}

	_, rateLimitData, err := o.RemoveMember(ctx, userEmail)
	var annos annotations.Annotations
	annos.WithRateLimiting(rateLimitData)

	if err != nil {
		l.Error("error deleting user", zap.Error(err), zap.String("email", userEmail))
		return annos, err
	}

	o.usersMutex.Lock()
	if o.users != nil {
		delete(o.users, accountID)
	}
	o.usersMutex.Unlock()

	return annos, nil
}
