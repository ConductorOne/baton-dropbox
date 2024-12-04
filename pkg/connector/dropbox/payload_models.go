package dropbox

// Users

type ListUsersPayload struct {
	Cursor  string        `json:"cursor"`
	HasMore bool          `json:"has_more"`
	Members []UserPayload `json:"members"`
}

type UserPayload struct {
	Profile Profile `json:"profile"`
	Roles   []Role  `json:"roles"`
}

type Profile struct {
	AccountID    string   `json:"account_id"`
	TeamMemberID string   `json:"team_member_id"`
	Name         Name     `json:"name"`
	Email        string   `json:"email"`
	Groups       []string `json:"groups"`
}

func (u UserPayload) HasRole(roleID string) bool {
	for _, role := range u.Roles {
		if role.RoleID == roleID {
			return true
		}
	}
	return false
}

type Name struct {
	DisplayName string `json:"display_name"`
	GivenName   string `json:"given_name"`
	Surname     string `json:"surname"`
}

// Roles

type Role struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	RoleID      string `json:"role_id"`
}

// Groups

type ListGroupsPayload struct {
	Cursor  string  `json:"cursor"`
	HasMore bool    `json:"has_more"`
	Groups  []Group `json:"groups"`
}

type Group struct {
	GroupID             string `json:"group_id"`
	Name                string `json:"group_name"`
	GroupManagementType Tag    `json:"group_management_type"`
	MemberCount         int    `json:"member_count"`
}

type Tag struct {
	Tag string `json:".tag"`
}

type ListGroupMembersPayload struct {
	Cursor  string           `json:"cursor"`
	HasMore bool             `json:"has_more"`
	Members []MembersPayload `json:"members"`
}

type MembersPayload struct {
	Profile MembersProfile `json:"profile"`

	// owner or member of the group
	AccessType Tag `json:"access_type"`
}

type MembersProfile struct {
	Profile
	TeamMemberID string `json:"team_member_id"`
}

type ListFoldersPayload struct {
	Cursor  string   `json:"cursor"`
	Entries []Folder `json:"entries"`
}

type Folder struct {
	SharedFolderId string `json:"shared_folder_id"`
	Name           string `json:"name"`
}

type ListFolderMembersPayload struct {
	Cursor string            `json:"cursor"`
	Groups []ListFolderGroup `json:"groups"`
	Users  []ListFolderUser  `json:"users"`
}

type ListFolderGroup struct {
	AccessType  Tag                   `json:"access_type"`
	Permissions []any                 `json:"permissions"`
	Group       ListFolderNestedGroup `json:"group"`
}

type ListFolderNestedGroup struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

type ListFolderUser struct {
	AccessType  Tag                  `json:"access_type"`
	Permissions []any                `json:"permissions"`
	User        ListFolderNestedUser `json:"user"`
}

type ListFolderNestedUser struct {
	AccountID    string `json:"account_id"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	TeamMemberID string `json:"team_member_id"`
}
