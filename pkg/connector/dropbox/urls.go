package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = "https://dropbox.com/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

// API Documentation: https://www.dropbox.com/developers/documentation/http/documentation

const (
	// User Management Endpoints
	// Documentation: https://www.dropbox.com/developers/documentation/http/teams#team-members

	// ListUsersURL lists team members
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-list
	// Required Scope: members.read.
	ListUsersURL = BaseURL + "/2/team/members/list_v2"

	// ListUsersContinueURL continues paginated user listing
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-list-continue
	// Required Scope: members.read.
	ListUsersContinueURL = BaseURL + "/2/team/members/list/continue_v2"

	// AddMemberURL provisions a new team member
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-add
	// Required Scope: members.write
	// Permission: Team member management.
	AddMemberURL = BaseURL + "/2/team/members/add_v2"

	// RemoveMemberURL removes a team member
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-remove
	// Required Scope: members.delete
	// Permission: Team member management.
	RemoveMemberURL = BaseURL + "/2/team/members/remove"

	// SuspendMemberURL suspends a team member
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-suspend
	// Required Scope: members.write
	// Permission: Team member management.
	SuspendMemberURL = BaseURL + "/2/team/members/suspend"

	// UnsuspendMemberURL reactivates a suspended team member
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-unsuspend
	// Required Scope: members.write
	// Permission: Team member management.
	UnsuspendMemberURL = BaseURL + "/2/team/members/unsuspend"

	// Role Management Endpoints
	// Documentation: https://www.dropbox.com/developers/documentation/http/teams#team-members-set_admin_permissions

	// SetRoleURL sets admin permissions for a team member
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-members-set_admin_permissions
	// Required Scope: members.write
	// Permission: Team member management.
	SetRoleURL = BaseURL + "/2/team/members/set_admin_permissions_v2"

	// Group Management Endpoints
	// Documentation: https://www.dropbox.com/developers/documentation/http/teams#team-groups

	// ListGroupsURL lists all groups
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-list
	// Required Scope: groups.read.
	ListGroupsURL = BaseURL + "/2/team/groups/list"

	// ListGroupsContinueURL continues paginated group listing
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-list-continue
	// Required Scope: groups.read.
	ListGroupsContinueURL = BaseURL + "/2/team/groups/list/continue"

	// ListGroupMembersURL lists members of a specific group
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-members-list
	// Required Scope: groups.read.
	ListGroupMembersURL = BaseURL + "/2/team/groups/members/list"

	// ListGroupMembersContinueURL continues paginated group member listing
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-members-list-continue
	// Required Scope: groups.read.
	ListGroupMembersContinueURL = BaseURL + "/2/team/groups/members/list/continue"

	// AddUserToGroupURL adds members to a group
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-members-add
	// Required Scope: groups.write
	// Permission: Team member management.
	AddUserToGroupURL = BaseURL + "/2/team/groups/members/add"

	// RemoveUserFromGroupURL removes members from a group
	// Docs: https://www.dropbox.com/developers/documentation/http/teams#team-groups-members-remove
	// Required Scope: groups.write
	// Permission: Team member management.
	RemoveUserFromGroupURL = BaseURL + "/2/team/groups/members/remove"
)
