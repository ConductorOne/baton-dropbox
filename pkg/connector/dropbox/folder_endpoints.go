package dropbox

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
//
// 	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
// 	"github.com/conductorone/baton-sdk/pkg/uhttp"
// )
//
// const folderDefaultLimit = 100
//
// type ListFoldersBody struct {
// 	IncludeDeleted bool   `json:"include_deleted"`
// 	Path           string `json:"path"`
// }
//
// func DefaultListFoldersBody() ListFoldersBody {
// 	return ListFoldersBody{
// 		IncludeDeleted: false,
// 		Path:           "",
// 	}
// }
//
// func (c *Client) ListFolders(ctx context.Context) (*ListFoldersPayload, *v2.RateLimitDescription, error) {
// 	body := DefaultListFoldersBody()
//
// 	reader := new(bytes.Buffer)
// 	err := json.NewEncoder(reader).Encode(body)
// 	if err != nil {
// 		return nil, nil, err
// 	}
//
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListFoldersURL, reader)
// 	if err != nil {
// 		return nil, nil, err
// 	}
//
// 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Dropbox-API-Select-Admin", "dbmid:AABZq4_Ono_sDDDq8imzm3jLaW1JXbA1xPE")
//
// 	var target ListFoldersPayload
// 	var ratelimitData v2.RateLimitDescription
// 	res, err := c.Do(req,
// 		uhttp.WithJSONResponse(&target),
// 		uhttp.WithRatelimitData(&ratelimitData),
// 	)
// 	if err != nil {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	defer res.Body.Close()
// 	if res.StatusCode != http.StatusOK {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	return &target, &ratelimitData, nil
// }
//
// func (c *Client) ListFoldersContinue(ctx context.Context, cursor string) (*ListFoldersPayload, *v2.RateLimitDescription, error) {
// 	body := struct {
// 		Cursor string `json:"cursor"`
// 	}{Cursor: cursor}
//
// 	reader := new(bytes.Buffer)
// 	err := json.NewEncoder(reader).Encode(body)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListFoldersContinueURL, reader)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Dropbox-API-Select-Admin", "dbmid:AABZq4_Ono_sDDDq8imzm3jLaW1JXbA1xPE")
//
// 	var target ListFoldersPayload
// 	var ratelimitData v2.RateLimitDescription
// 	res, err := c.Do(req,
// 		uhttp.WithJSONResponse(&target),
// 		uhttp.WithRatelimitData(&ratelimitData),
// 	)
// 	if err != nil {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	defer res.Body.Close()
// 	if res.StatusCode != http.StatusOK {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	return &target, &ratelimitData, nil
// }
//
// type ListFolderMembersBody struct {
// 	Actions        any    `json:"actions"`
// 	Limit          int    `json:"limit"`
// 	SharedFolderId string `json:"shared_folder_id"`
// }
//
// func DefaultFolderMembersBody() ListFolderMembersBody {
// 	return ListFolderMembersBody{
// 		Limit: folderDefaultLimit,
// 	}
// }
//
// func (c *Client) ListFolderMembers(ctx context.Context, sharedFolderId string, limit int) (*ListFolderMembersPayload, *v2.RateLimitDescription, error) {
// 	body := DefaultFolderMembersBody()
// 	if sharedFolderId == "" {
// 		return nil, nil, fmt.Errorf("sharedFolderId is required")
// 	}
// 	body.SharedFolderId = sharedFolderId
//
// 	if limit != 0 {
// 		body.Limit = limit
// 	}
//
// 	reader := new(bytes.Buffer)
// 	err := json.NewEncoder(reader).Encode(body)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListFolderMembersURL, reader)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// 	req.Header.Set("Content-Type", "application/json")
//
// 	var target ListFolderMembersPayload
// 	var ratelimitData v2.RateLimitDescription
// 	res, err := c.Do(req,
// 		uhttp.WithJSONResponse(&target),
// 		uhttp.WithRatelimitData(&ratelimitData),
// 	)
//
// 	if err != nil {
// 		return nil, &ratelimitData, err
// 	}
//
// 	defer res.Body.Close()
// 	if res.StatusCode != http.StatusOK {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	return &target, &ratelimitData, nil
// }
//
// func (c *Client) ListFolderMembersContinue(ctx context.Context, cursor string) (*ListFolderMembersPayload, *v2.RateLimitDescription, error) {
// 	body := struct {
// 		Cursor string `json:"cursor"`
// 	}{Cursor: cursor}
//
// 	reader := new(bytes.Buffer)
// 	err := json.NewEncoder(reader).Encode(body)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ListFolderMembersContinueURL, reader)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// 	req.Header.Set("Content-Type", "application/json")
//
// 	var target ListFolderMembersPayload
// 	var ratelimitData v2.RateLimitDescription
// 	res, err := c.Do(req,
// 		uhttp.WithJSONResponse(&target),
// 		uhttp.WithRatelimitData(&ratelimitData),
// 	)
//
// 	if err != nil {
// 		return nil, &ratelimitData, err
// 	}
//
// 	defer res.Body.Close()
// 	if res.StatusCode != http.StatusOK {
// 		logBody(ctx, res.Body)
// 		return nil, &ratelimitData, err
// 	}
//
// 	return &target, &ratelimitData, nil
// }
//
// //
// // type RemoveUserFromFolderBody struct {
// // 	Folder        FolderIdTag `json:"group"`
// // 	Users         []EmailTag  `json:"users"`
// // 	ReturnMembers bool        `json:"return_members"`
// // }
// //
// // func (c *Client) RemoveUserFromFolder(ctx context.Context, groupId, email string) (*v2.RateLimitDescription, error) {
// // 	body := RemoveUserFromFolderBody{
// // 		Folder: FolderIdTag{
// // 			FolderID: groupId,
// // 			Tag:      "group_id",
// // 		},
// // 		Users: []EmailTag{
// // 			{
// // 				Tag:   "email",
// // 				Email: email,
// // 			},
// // 		},
// // 	}
// //
// // 	buffer := new(bytes.Buffer)
// // 	err := json.NewEncoder(buffer).Encode(body)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RemoveUserFromFolderURL, buffer)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// // 	req.Header.Set("Content-Type", "application/json")
// //
// // 	var ratelimitData v2.RateLimitDescription
// // 	res, err := c.Do(req,
// // 		uhttp.WithRatelimitData(&ratelimitData),
// // 	)
// //
// // 	if err != nil {
// // 		logBody(ctx, res.Body)
// // 		return &ratelimitData, err
// // 	}
// //
// // 	defer res.Body.Close()
// // 	if res.StatusCode != http.StatusOK {
// // 		logBody(ctx, res.Body)
// // 		return &ratelimitData, err
// // 	}
// //
// // 	return &ratelimitData, nil
// // }
// //
// // type AddUserToFolderBody struct {
// // 	Folder        FolderIdTag          `json:"group"`
// // 	Members       []AddToFolderMembers `json:"members"`
// // 	ReturnMembers bool                 `json:"return_members"`
// // }
// //
// // type AddToFolderMembers struct {
// // 	AccessLevel string   `json:"access_type"`
// // 	User        EmailTag `json:"user"`
// // }
// //
// // func (c *Client) AddUserToFolder(ctx context.Context, groupId, email, accessType string) (*v2.RateLimitDescription, error) {
// // 	body := AddUserToFolderBody{
// // 		Folder: FolderIdTag{
// // 			Tag:      "group_id",
// // 			FolderID: groupId,
// // 		},
// // 		Members: []AddToFolderMembers{
// // 			{
// // 				AccessLevel: accessType,
// // 				User: EmailTag{
// // 					Tag:   "email",
// // 					Email: email,
// // 				},
// // 			},
// // 		},
// // 	}
// //
// // 	buf := new(bytes.Buffer)
// // 	err := json.NewEncoder(buf).Encode(body)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// //
// // 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AddUserToFolderURL, buf)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
// // 	req.Header.Set("Content-Type", "application/json")
// //
// // 	var ratelimitData v2.RateLimitDescription
// // 	res, err := c.Do(req,
// // 		uhttp.WithRatelimitData(&ratelimitData),
// // 	)
// // 	if err != nil {
// // 		logBody(ctx, res.Body)
// // 		return &ratelimitData, err
// // 	}
// //
// // 	defer res.Body.Close()
// // 	if res.StatusCode != http.StatusOK {
// // 		logBody(ctx, res.Body)
// // 		return &ratelimitData, err
// // 	}
// //
// // 	return &ratelimitData, nil
// // }
