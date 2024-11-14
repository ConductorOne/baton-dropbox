package dropbox

import (
	"net/http"
)

type Client struct {
	http.Client
	AccessToken string
}
