package connector

import (
	"context"
	"fmt"

	config "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	ActionSuspendUser   = "suspend_user"
	ActionUnsuspendUser = "unsuspend_user"
)

var suspendUserActionSchema = &v2.BatonActionSchema{
	Name:        ActionSuspendUser,
	DisplayName: "Suspend User",
	Description: "Suspends a user's access to Dropbox Team",
	Arguments: []*config.Field{
		{
			Name:        "user_id",
			DisplayName: "User Account ID",
			Description: "The account ID of the user to suspend",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Description: "Whether the user was suspended successfully",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
	},
}

var unsuspendUserActionSchema = &v2.BatonActionSchema{
	Name:        ActionUnsuspendUser,
	DisplayName: "Unsuspend User",
	Description: "Unsuspends (reactivates) a user's access to Dropbox Team",
	Arguments: []*config.Field{
		{
			Name:        "user_id",
			DisplayName: "User Account ID",
			Description: "The account ID of the user to unsuspend",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Description: "Whether the user was unsuspended successfully",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
	},
}

// RegisterActionManager registers custom actions for the Dropbox connector.
func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, suspendUserActionSchema.Name, suspendUserActionSchema, c.suspendUserActionHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to register suspend user action: %w", err)
	}

	err = actionManager.RegisterAction(ctx, unsuspendUserActionSchema.Name, unsuspendUserActionSchema, c.unsuspendUserActionHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to register unsuspend user action: %w", err)
	}

	return actionManager, nil
}

// suspendUserActionHandler handles the suspend user action.
func (c *Connector) suspendUserActionHandler(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if args == nil || args.Fields == nil {
		l.Error("baton-dropbox: suspend user action error - invalid arguments")
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid arguments")
	}

	resourceIDValue, exists := args.Fields["user_id"]
	if !exists || resourceIDValue == nil {
		l.Error("baton-dropbox: suspend user action error - missing user ID")
		return nil, nil, status.Errorf(codes.InvalidArgument, "missing user_id")
	}

	uidField, ok := resourceIDValue.GetKind().(*structpb.Value_StringValue)
	if !ok {
		l.Error("baton-dropbox: suspend user action error - invalid user ID format")
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid user_id format")
	}

	accountID := uidField.StringValue
	if accountID == "" {
		l.Error("baton-dropbox: suspend user action error - empty user ID")
		return nil, nil, status.Errorf(codes.InvalidArgument, "user_id cannot be empty")
	}

	l.Info("baton-dropbox: suspending user", zap.String("account_id", accountID))

	err := c.suspendUser(ctx, accountID)
	if err != nil {
		l.Error("baton-dropbox: failed to suspend user", zap.String("account_id", accountID), zap.Error(err))
		return nil, nil, fmt.Errorf("baton-dropbox: failed to suspend user: %w", err)
	}

	l.Info("baton-dropbox: user suspended successfully", zap.String("account_id", accountID))

	return getResponseStruct(true), nil, nil
}

// unsuspendUserActionHandler handles the unsuspend user action.
func (c *Connector) unsuspendUserActionHandler(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if args == nil || args.Fields == nil {
		l.Error("baton-dropbox: unsuspend user action error - invalid arguments")
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid arguments")
	}

	resourceIDValue, exists := args.Fields["user_id"]
	if !exists || resourceIDValue == nil {
		l.Error("baton-dropbox: unsuspend user action error - missing user ID")
		return nil, nil, status.Errorf(codes.InvalidArgument, "missing user_id")
	}

	uidField, ok := resourceIDValue.GetKind().(*structpb.Value_StringValue)
	if !ok {
		l.Error("baton-dropbox: unsuspend user action error - invalid user ID format")
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid user_id format")
	}

	accountID := uidField.StringValue
	if accountID == "" {
		l.Error("baton-dropbox: unsuspend user action error - empty user ID")
		return nil, nil, status.Errorf(codes.InvalidArgument, "user_id cannot be empty")
	}

	l.Info("baton-dropbox: unsuspending user", zap.String("account_id", accountID))

	err := c.unsuspendUser(ctx, accountID)
	if err != nil {
		l.Error("baton-dropbox: failed to unsuspend user", zap.String("account_id", accountID), zap.Error(err))
		return nil, nil, fmt.Errorf("baton-dropbox: failed to unsuspend user: %w", err)
	}

	l.Info("baton-dropbox: user unsuspended successfully", zap.String("account_id", accountID))

	return getResponseStruct(true), nil, nil
}

// getResponseStruct creates a standard response struct for action results.
func getResponseStruct(success bool) *structpb.Struct {
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": {
				Kind: &structpb.Value_BoolValue{BoolValue: success},
			},
		},
	}
}

// suspendUser suspends a user by their account ID.
func (c *Connector) suspendUser(ctx context.Context, accountID string) error {
	l := ctxzap.Extract(ctx)

	// Get user email from account ID using the cache
	userBuilder := newUserBuilder(c.client)
	email, err := userBuilder.getEmailByAccountID(ctx, accountID)
	if err != nil {
		l.Error("baton-dropbox: failed to get email for user", zap.String("account_id", accountID), zap.Error(err))
		return fmt.Errorf("failed to get email for account_id %s: %w", accountID, err)
	}

	// Call the suspend endpoint
	rateLimitData, err := c.client.SuspendMember(ctx, email)
	if err != nil {
		l.Error("baton-dropbox: failed to suspend member", zap.String("email", email), zap.Error(err))
		return fmt.Errorf("failed to suspend member %s: %w", email, err)
	}

	// Log rate limit info if present
	if rateLimitData != nil {
		l.Debug("baton-dropbox: suspend user rate limit info",
			zap.String("email", email),
			zap.Int64("limit", rateLimitData.Limit),
			zap.Int64("remaining", rateLimitData.Remaining),
		)
	}

	return nil
}

// unsuspendUser unsuspends (reactivates) a user by their account ID.
func (c *Connector) unsuspendUser(ctx context.Context, accountID string) error {
	l := ctxzap.Extract(ctx)

	// Get user email from account ID using the cache
	userBuilder := newUserBuilder(c.client)
	email, err := userBuilder.getEmailByAccountID(ctx, accountID)
	if err != nil {
		l.Error("baton-dropbox: failed to get email for user", zap.String("account_id", accountID), zap.Error(err))
		return fmt.Errorf("failed to get email for account_id %s: %w", accountID, err)
	}

	// Call the unsuspend endpoint
	rateLimitData, err := c.client.UnsuspendMember(ctx, email)
	if err != nil {
		l.Error("baton-dropbox: failed to unsuspend member", zap.String("email", email), zap.Error(err))
		return fmt.Errorf("failed to unsuspend member %s: %w", email, err)
	}

	// Log rate limit info if present
	if rateLimitData != nil {
		l.Debug("baton-dropbox: unsuspend user rate limit info",
			zap.String("email", email),
			zap.Int64("limit", rateLimitData.Limit),
			zap.Int64("remaining", rateLimitData.Remaining),
		)
	}

	return nil
}
