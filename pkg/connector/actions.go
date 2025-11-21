package connector

import (
	"context"
	"fmt"
	"strings"

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
	ActionDisableUser = "disable_user"
	ActionEnableUser  = "enable_user"
)

var disableUserActionSchema = &v2.BatonActionSchema{
	Name:        ActionDisableUser,
	DisplayName: "Disable User",
	Description: "Disables a user's access to Dropbox Team (suspends the account)",
	Arguments: []*config.Field{
		{
			Name:        "user_id",
			DisplayName: "User Team Member ID",
			Description: "The team member ID of the user to disable",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Description: "Whether the user was disabled successfully",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT_DISABLE,
	},
}

var enableUserActionSchema = &v2.BatonActionSchema{
	Name:        ActionEnableUser,
	DisplayName: "Enable User",
	Description: "Enables a user's access to Dropbox Team (unsuspends the account)",
	Arguments: []*config.Field{
		{
			Name:        "user_id",
			DisplayName: "User Team Member ID",
			Description: "The team member ID of the user to enable",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Description: "Whether the user was enabled successfully",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT_ENABLE,
	},
}

// extractUserID extracts and validates the user_id from action arguments.
func extractUserID(ctx context.Context, args *structpb.Struct, actionName string) (string, error) {
	l := ctxzap.Extract(ctx)

	if args == nil || args.Fields == nil {
		l.Error("invalid arguments", zap.String("action", actionName))
		return "", status.Errorf(codes.InvalidArgument, "invalid arguments")
	}

	resourceIDValue, exists := args.Fields["user_id"]
	if !exists || resourceIDValue == nil {
		l.Error("missing user ID", zap.String("action", actionName))
		return "", status.Errorf(codes.InvalidArgument, "missing user_id")
	}

	uidField, ok := resourceIDValue.GetKind().(*structpb.Value_StringValue)
	if !ok {
		l.Error("invalid user ID format", zap.String("action", actionName))
		return "", status.Errorf(codes.InvalidArgument, "invalid user_id format")
	}

	teamMemberID := uidField.StringValue
	if teamMemberID == "" {
		l.Error("empty user ID", zap.String("action", actionName))
		return "", status.Errorf(codes.InvalidArgument, "user_id cannot be empty")
	}

	return teamMemberID, nil
}

// RegisterActionManager registers custom actions for the Dropbox connector.
func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, disableUserActionSchema.Name, disableUserActionSchema, c.disableUserActionHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to register disable user action: %w", err)
	}

	err = actionManager.RegisterAction(ctx, enableUserActionSchema.Name, enableUserActionSchema, c.enableUserActionHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to register enable user action: %w", err)
	}

	return actionManager, nil
}

// disableUserActionHandler handles the disable user action.
func (c *Connector) disableUserActionHandler(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	teamMemberID, err := extractUserID(ctx, args, ActionDisableUser)
	if err != nil {
		return nil, nil, err
	}

	l.Info("disabling user", zap.String("team_member_id", teamMemberID))

	_, err = c.client.SuspendMember(ctx, teamMemberID)
	if err != nil {
		if strings.Contains(err.Error(), "suspend_inactive_user") {
			l.Info("user is already disabled", zap.String("team_member_id", teamMemberID))
			return getResponseStruct(true), nil, nil
		}
		l.Error("failed to disable user", zap.String("team_member_id", teamMemberID), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to disable user: %w", err)
	}

	l.Info("user disabled successfully", zap.String("team_member_id", teamMemberID))
	return getResponseStruct(true), nil, nil
}

// enableUserActionHandler handles the enable user action.
func (c *Connector) enableUserActionHandler(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	teamMemberID, err := extractUserID(ctx, args, ActionEnableUser)
	if err != nil {
		return nil, nil, err
	}

	l.Info("enabling user", zap.String("team_member_id", teamMemberID))

	_, err = c.client.UnsuspendMember(ctx, teamMemberID)
	if err != nil {
		if strings.Contains(err.Error(), "unsuspend_non_suspended_member") {
			l.Info("user is already enabled", zap.String("team_member_id", teamMemberID))
			return getResponseStruct(true), nil, nil
		}
		l.Error("failed to enable user", zap.String("team_member_id", teamMemberID), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to enable user: %w", err)
	}

	l.Info("user enabled successfully", zap.String("team_member_id", teamMemberID))
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
