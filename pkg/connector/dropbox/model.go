package dropbox

// ListUsersPayload represents the top-level JSON structure.
type ListUsersPayload struct {
	Cursor  string   `json:"cursor"`
	HasMore bool     `json:"has_more"`
	Members []Member `json:"members"`
}

// Member represents an individual member in the "members" array.
type Member struct {
	Profile User   `json:"profile"`
	Roles   []Role `json:"roles"`
}

// User represents the basic personal information and groups of a member.
type User struct {
	AccountID string   `json:"account_id"`
	Name      Name     `json:"name"`
	Email     string   `json:"email"`
	Groups    []string `json:"groups"`
	// MembershipType string   `json:"membership_type,omitempty"` // .tag value as string
	// Status         string   `json:"status,omitempty"`          // .tag value as string
}

// Name represents the "name" object in the profile.
type Name struct {
	DisplayName string `json:"display_name"`
	GivenName   string `json:"given_name"`
	Surname     string `json:"surname"`
}

// Role represents the roles assigned to a member.
type Role struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	RoleID      string `json:"role_id"`
}
