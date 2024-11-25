package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = "https://dropbox.com/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

const (
	ListUsersURL         = BaseURL + "/2/team/members/list_v2"
	ListUsersContinueURL = BaseURL + "/2/team/members/list/continue_v2"
	AddRoleToUserURL     = BaseURL + "/2/team/members/add_role"
)

const (
	ListGroupsURL               = BaseURL + "/2/team/groups/list"
	ListGroupMembersURL         = BaseURL + "/2/team/groups/members/list"
	AddGroupAccessTypeToUserURL = BaseURL + "/2/team/groups/members/set_access_type"
)
