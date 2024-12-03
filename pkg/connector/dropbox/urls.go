package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = "https://dropbox.com/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

const (
	ListUsersURL         = BaseURL + "/2/team/members/list_v2"
	ListUsersContinueURL = BaseURL + "/2/team/members/list/continue_v2"
)

const (
	AddRoleToUserURL = BaseURL + "/2/team/members/set_admin_permissions_v2"
)

const (
	ListGroupsURL               = BaseURL + "/2/team/groups/list"
	ListGroupsContinueURL       = BaseURL + "/2/team/groups/list/continue"
	ListGroupMembersURL         = BaseURL + "/2/team/groups/members/list"
	ListGroupMembersContinueURL = BaseURL + "/2/team/groups/members/list/continue"
	AddGroupAccessTypeToUserURL = BaseURL + "/2/team/groups/members/set_access_type"
	AddUserToGroupURL           = BaseURL + "/2/team/groups/members/add"
	RemoveUserFromGroupURL      = BaseURL + "/2/team/groups/members/remove"
)
