package dropbox

// Common Types

// Tag represents a Dropbox API discriminator field used to indicate object types.
type Tag struct {
	Tag string `json:".tag"`
}

// DropboxError represents the standard error response from Dropbox API.
// Implements the uhttp.ErrorResponse interface.
type DropboxError struct {
	ErrorSummary string `json:"error_summary"`
	Error        struct {
		Tag string `json:".tag"`
	} `json:"error"`
}

// Message implements the ErrorResponse interface required by uhttp.WithErrorResponse.
func (e *DropboxError) Message() string {
	if e.ErrorSummary != "" {
		return e.ErrorSummary
	}
	if e.Error.Tag != "" {
		return e.Error.Tag
	}
	return "unknown dropbox error"
}

// EmailTag represents an email-based identifier used in Dropbox API requests.
type EmailTag struct {
	Tag   string `json:".tag"`
	Email string `json:"email"`
}

// TeamMemberIdTag represents a team_member_id-based identifier used in Dropbox API requests.
type TeamMemberIdTag struct {
	Tag          string `json:".tag"`
	TeamMemberID string `json:"team_member_id"`
}

// GroupIdTag represents a group identifier used in Dropbox API requests.
type GroupIdTag struct {
	GroupID string `json:"group_id"`
	Tag     string `json:".tag"`
}

// Name represents a user's full name in Dropbox.
type Name struct {
	DisplayName string `json:"display_name"`
	GivenName   string `json:"given_name"`
	Surname     string `json:"surname"`
}

// Users

// ListUsersPayload represents the response from the list users API endpoint.
type ListUsersPayload struct {
	Cursor  string        `json:"cursor"`
	HasMore bool          `json:"has_more"`
	Members []UserPayload `json:"members"`
}

// ListUserBody represents the request body for listing users.
type ListUserBody struct {
	Limit          int  `json:"limit"`
	IncludeRemoved bool `json:"include_removed"`
}

// UserPayload represents a team member with their profile and roles.
type UserPayload struct {
	Profile Profile `json:"profile"`
	Roles   []Role  `json:"roles"`
}

// Profile represents a user's profile information in Dropbox.
type Profile struct {
	AccountID    string   `json:"account_id"`
	TeamMemberID string   `json:"team_member_id"`
	Name         Name     `json:"name"`
	Email        string   `json:"email"`
	Groups       []string `json:"groups"`
	Status       Tag      `json:"status"`
}

// Last Login (Team Audit Log)

// EventTypeArg is the union selector used to filter get_events by a single event type.
// For last-login we filter on "login_success".
type EventTypeArg struct {
	Tag string `json:".tag"`
}

// GetEventsBody represents the request body for /2/team_log/get_events.
// AccountID filters to events where the member is the Actor, Context, or Participant.
type GetEventsBody struct {
	Limit     int           `json:"limit"`
	AccountID string        `json:"account_id,omitempty"`
	EventType *EventTypeArg `json:"event_type,omitempty"`
}

// GetEventsPayload represents the response from /2/team_log/get_events.
type GetEventsPayload struct {
	Events  []EventEntry `json:"events"`
	Cursor  string       `json:"cursor"`
	HasMore bool         `json:"has_more"`
}

// EventEntry is a single audit-log event. Timestamp is RFC3339 (e.g. "2020-01-01T00:00:00Z").
type EventEntry struct {
	Timestamp string `json:"timestamp"`
	EventType Tag    `json:"event_type"`
}

// Account Provisioning

// AddMemberRequest represents the request body for adding team members.
type AddMemberRequest struct {
	NewMembers []NewMemberInfo `json:"new_members"`
}

// NewMemberInfo represents information for a new team member to be added.
type NewMemberInfo struct {
	MemberEmail string `json:"member_email"`
}

// AddMemberResponse represents the response from adding team members.
type AddMemberResponse struct {
	Tag      string            `json:".tag"`
	Complete []AddMemberResult `json:"complete"`
}

// AddMemberResult represents the result of adding a single member.
type AddMemberResult struct {
	Tag     string  `json:".tag"`
	Profile Profile `json:"profile,omitempty"`
}

// Account Deprovisioning

// RemoveMemberRequest represents the request body for removing a team member.
type RemoveMemberRequest struct {
	User TeamMemberIdTag `json:"user"`
}

// RemoveMemberResponse represents the response from removing a team member.
type RemoveMemberResponse struct {
	Tag string `json:".tag"`
}

// Roles

// Role represents a team role in Dropbox.
type Role struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	RoleID      string `json:"role_id"`
}

// addRoleToUserBody represents the request body for role operations.
type addRoleToUserBody struct {
	NewRoles   []string        `json:"new_roles"`
	TeamMember TeamMemberIdTag `json:"user"`
}

// Groups

// ListGroupsPayload represents the response from the list groups API endpoint.
type ListGroupsPayload struct {
	Cursor  string  `json:"cursor"`
	HasMore bool    `json:"has_more"`
	Groups  []Group `json:"groups"`
}

// ListGroupsBody represents the request body for listing groups.
type ListGroupsBody struct {
	Limit int `json:"limit"`
}

// Group represents a team group in Dropbox.
type Group struct {
	GroupID             string `json:"group_id"`
	Name                string `json:"group_name"`
	GroupManagementType Tag    `json:"group_management_type"`
	MemberCount         int    `json:"member_count"`
}

// ListGroupMembersPayload represents the response from the list group members API endpoint.
type ListGroupMembersPayload struct {
	Cursor  string           `json:"cursor"`
	HasMore bool             `json:"has_more"`
	Members []MembersPayload `json:"members"`
}

// ListGroupMembersBody represents the request body for listing group members.
type ListGroupMembersBody struct {
	Group GroupIdTag `json:"group"`
	Limit int        `json:"limit"`
}

// MembersPayload represents a group member with their profile and access type.
type MembersPayload struct {
	Profile    MembersProfile `json:"profile"`
	AccessType Tag            `json:"access_type"` // "owner" or "member"
}

// MembersProfile represents a group member's profile (extends Profile).
type MembersProfile struct {
	Profile
}

// RemoveUserFromGroupBody represents the request body for removing users from a group.
type RemoveUserFromGroupBody struct {
	Group         GroupIdTag        `json:"group"`
	Users         []TeamMemberIdTag `json:"users"`
	ReturnMembers bool              `json:"return_members"`
}

// AddUserToGroupBody represents the request body for adding users to a group.
type AddUserToGroupBody struct {
	Group         GroupIdTag          `json:"group"`
	Members       []AddToGroupMembers `json:"members"`
	ReturnMembers bool                `json:"return_members"`
}

// AddToGroupMembers represents a member to be added to a group with access level.
type AddToGroupMembers struct {
	AccessLevel Tag             `json:"access_type"` // union tag: "member" or "owner"
	User        TeamMemberIdTag `json:"user"`
}
