package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = "https://dropbox.com/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

// https://www.dropbox.com/developers/documentation/http/documentation

const (
	// users.
	ListUsersURL         = BaseURL + "/2/team/members/list_v2"
	ListUsersContinueURL = BaseURL + "/2/team/members/list/continue_v2"

	// roles.
	SetRoleURL = BaseURL + "/2/team/members/set_admin_permissions_v2"

	// groups.
	ListGroupsURL               = BaseURL + "/2/team/groups/list"
	ListGroupsContinueURL       = BaseURL + "/2/team/groups/list/continue"
	ListGroupMembersURL         = BaseURL + "/2/team/groups/members/list"
	ListGroupMembersContinueURL = BaseURL + "/2/team/groups/members/list/continue"
	AddUserToGroupURL           = BaseURL + "/2/team/groups/members/add"
	RemoveUserFromGroupURL      = BaseURL + "/2/team/groups/members/remove"
)
