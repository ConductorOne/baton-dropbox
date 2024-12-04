package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = "https://dropbox.com/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

// user endpoints.
const (
	ListFoldersURL               = BaseURL + "/2/files/list_folder"
	ListFoldersContinueURL       = BaseURL + "/2/files/list_folder/continue"
	AddUserToFolderURL           = BaseURL + "/2/sharing/add_folder_member"
	RemoveUserFromFolderURL      = BaseURL + "/2/sharing/remove_folder_member"
	ListFolderMembersURL         = BaseURL + "/2/sharing/list_folder_members"
	ListFolderMembersContinueURL = BaseURL + "/2/sharing/list_folder_members/continue"
)

// enterprise endpoints.
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
