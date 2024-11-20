package dropbox

// Users

type ListUsersPayload struct {
	Cursor  string           `json:"cursor"`
	HasMore bool             `json:"has_more"`
	Members []ListUserMember `json:"members"`
}

type ListUserMember struct {
	Profile User   `json:"profile"`
	Roles   []Role `json:"roles"`
}

type User struct {
	AccountID string   `json:"account_id"`
	Name      Name     `json:"name"`
	Email     string   `json:"email"`
	Groups    []string `json:"groups"`
}

func (u ListUserMember) HasRole(roleID string) bool {
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
	GroupID             string              `json:"group_id"`
	Name                string              `json:"group_name"`
	GroupManagementType GroupManagementType `json:"group_management_type"`
	MemberCount         int                 `json:"member_count"`
}

type GroupManagementType struct {
	Tag string `json:".tag"`
}

type ListGroupMembersPayload struct {
	Cursor  string            `json:"cursor"`
	HasMore bool              `json:"has_more"`
	Members []ListGroupMember `json:"members"`
}

type ListGroupMember struct {
	Profile User `json:"profile"`

	// owner or member of the group
	AccessType string `json:"access_type"`
}
