package dropbox

const (
	BaseURL  = "https://api.dropboxapi.com"
	AuthURL  = BaseURL + "/oauth2/authorize"
	TokenURL = BaseURL + "/oauth2/token"
)

const (
	ListUsersURL = BaseURL + "/2/team/members/list_v2"
)
