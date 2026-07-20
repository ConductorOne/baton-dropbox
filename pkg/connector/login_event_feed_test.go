package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLoginEventPageToken_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := &loginEventPageToken{
		LatestEventSeen: "2024-01-02T03:04:05Z",
		NextPageToken:   "next-cursor",
		StartAt:         "2024-01-01T00:00:00Z",
	}

	encoded, err := original.marshal()
	require.NoError(t, err)

	decoded, err := unmarshalLoginEventPageToken(&pagination.StreamToken{Cursor: encoded}, nil)
	require.NoError(t, err)
	require.Equal(t, original, decoded)
}

func TestLoginEventPageToken_DefaultsWhenEmpty(t *testing.T) {
	defaultStart := timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	pt, err := unmarshalLoginEventPageToken(nil, defaultStart)
	require.NoError(t, err)
	require.Equal(t, "2024-01-01T00:00:00Z", pt.StartAt)
	require.Equal(t, pt.StartAt, pt.LatestEventSeen)
	require.Empty(t, pt.NextPageToken)
}

func TestLoginEventPageToken_DefaultsToCatchUpWindowWhenNoStartGiven(t *testing.T) {
	pt, err := unmarshalLoginEventPageToken(nil, nil)
	require.NoError(t, err)

	startAt, err := time.Parse(dropbox.TimestampFormat, pt.StartAt)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(-defaultCatchUpWindow), startAt, time.Minute)
}

// newTestLoginEventFeed points a real dropbox.Client at an httptest server so
// ListEvents can be exercised end-to-end (request building, response parsing,
// filtering, cursor progression) without an interface seam over the client.
func newTestLoginEventFeed(t *testing.T, server *httptest.Server) *loginEventFeed {
	t.Helper()

	client, err := dropbox.NewClient(context.Background(), dropbox.Config{BaseURL: server.URL})
	require.NoError(t, err)
	client.TokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})

	return newLoginEventFeed(client)
}

func TestLoginEventFeed_ListEvents_FiltersToSuccessfulUserLogins(t *testing.T) {
	payload := dropbox.GetTeamEventsPayload{
		Events: []dropbox.TeamEvent{
			{
				Timestamp:     "2024-01-01T10:00:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor: dropbox.ActorLogInfo{
					Tag: "user",
					User: &dropbox.UserLogInfo{
						TeamMemberID: "dbmid:1",
						Email:        "alice@example.com",
						DisplayName:  "Alice",
					},
				},
			},
			{
				// A team member with admin permissions is logged under the "admin"
				// actor variant instead of "user" (same UserLogInfo shape) — this
				// must still be treated as a resolvable user login.
				Timestamp:     "2024-01-01T10:02:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor: dropbox.ActorLogInfo{
					Tag: "admin",
					Admin: &dropbox.UserLogInfo{
						TeamMemberID: "dbmid:3",
						Email:        "carol-admin@example.com",
						DisplayName:  "Carol",
					},
				},
			},
			{
				// Wrong event type: same category, but not a successful login.
				Timestamp:     "2024-01-01T10:05:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_fail"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{TeamMemberID: "dbmid:2"},
				},
			},
			{
				// Non-resolvable actor (system/Dropbox-initiated): should be skipped.
				Timestamp:     "2024-01-01T10:10:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor:         dropbox.ActorLogInfo{Tag: "dropbox"},
			},
			{
				// Missing team_member_id: not resolvable to a synced user, should be skipped.
				Timestamp:     "2024-01-01T10:15:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{Email: "no-id@example.com"},
				},
			},
		},
		Cursor:  "",
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/2/team_log/get_events", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(payload))
	}))
	defer server.Close()

	feed := newTestLoginEventFeed(t, server)

	events, streamState, _, err := feed.ListEvents(context.Background(), nil, nil)
	require.NoError(t, err)
	require.False(t, streamState.HasMore)
	require.Len(t, events, 2)

	usageEvent := events[0].GetUsageEvent()
	require.NotNil(t, usageEvent)
	require.Equal(t, "dbmid:1", usageEvent.ActorResource.Id.Resource)
	require.Equal(t, userResourceType.Id, usageEvent.ActorResource.Id.ResourceType)
	require.Equal(t, "Alice", usageEvent.ActorResource.DisplayName)

	adminActorEvent := events[1].GetUsageEvent()
	require.NotNil(t, adminActorEvent)
	require.Equal(t, "dbmid:3", adminActorEvent.ActorResource.Id.Resource)
	require.Equal(t, userResourceType.Id, adminActorEvent.ActorResource.Id.ResourceType)
	require.Equal(t, "Carol", adminActorEvent.ActorResource.DisplayName)
	require.Equal(t, dropboxAppResourceID, usageEvent.TargetResource.Id.Resource)
	require.Equal(t, appResourceType.Id, usageEvent.TargetResource.Id.ResourceType)
	require.Equal(t, timestamppb.New(mustParseDropboxTime(t, "2024-01-01T10:00:00Z")).AsTime(), events[0].OccurredAt.AsTime())
}

func TestLoginEventFeed_ListEvents_PaginatesUntilHasMoreFalse(t *testing.T) {
	firstPage := dropbox.GetTeamEventsPayload{
		Events: []dropbox.TeamEvent{
			{
				Timestamp:     "2024-01-01T10:00:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{TeamMemberID: "dbmid:1", Email: "alice@example.com"},
				},
			},
		},
		Cursor:  "page-2-cursor",
		HasMore: true,
	}
	secondPage := dropbox.GetTeamEventsPayload{
		Events: []dropbox.TeamEvent{
			{
				Timestamp:     "2024-01-01T11:00:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_success"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{TeamMemberID: "dbmid:2", Email: "bob@example.com"},
				},
			},
		},
		Cursor:  "",
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/2/team_log/get_events":
			require.NoError(t, json.NewEncoder(w).Encode(firstPage))
		case "/2/team_log/get_events/continue":
			var body dropbox.GetTeamEventsContinueBody
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "page-2-cursor", body.Cursor)
			require.NoError(t, json.NewEncoder(w).Encode(secondPage))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	feed := newTestLoginEventFeed(t, server)

	events, streamState, _, err := feed.ListEvents(context.Background(), nil, nil)
	require.NoError(t, err)
	require.True(t, streamState.HasMore)
	require.Len(t, events, 1)
	require.Equal(t, "dbmid:1", events[0].GetUsageEvent().ActorResource.Id.Resource)

	events, streamState, _, err = feed.ListEvents(context.Background(), nil, &pagination.StreamToken{Cursor: streamState.Cursor})
	require.NoError(t, err)
	require.False(t, streamState.HasMore)
	require.Len(t, events, 1)
	require.Equal(t, "dbmid:2", events[0].GetUsageEvent().ActorResource.Id.Resource)
}

func TestLoginEventFeed_ListEvents_AdvancesStartAtOnQuietTeam(t *testing.T) {
	// A window containing login-category events but no successful logins (only
	// logouts/failures). No usage events are emitted, but the high-water mark
	// must still advance to the newest event so the next sync's StartAt moves
	// forward instead of re-scanning a growing window.
	payload := dropbox.GetTeamEventsPayload{
		Events: []dropbox.TeamEvent{
			{
				Timestamp:     "2024-01-01T10:00:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "login_fail"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{TeamMemberID: "dbmid:1"},
				},
			},
			{
				Timestamp:     "2024-01-01T10:30:00Z",
				EventCategory: dropbox.Tag{Tag: "logins"},
				EventType:     dropbox.Tag{Tag: "logout"},
				Actor: dropbox.ActorLogInfo{
					Tag:  "user",
					User: &dropbox.UserLogInfo{TeamMemberID: "dbmid:2"},
				},
			},
		},
		Cursor:  "",
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(payload))
	}))
	defer server.Close()

	feed := newTestLoginEventFeed(t, server)

	startAt := timestamppb.New(mustParseDropboxTime(t, "2024-01-01T00:00:00Z"))
	events, streamState, _, err := feed.ListEvents(context.Background(), startAt, nil)
	require.NoError(t, err)
	require.Empty(t, events)
	require.False(t, streamState.HasMore)

	// StartAt for the next sync must advance to the newest event seen
	// (the logout at 10:30), not remain pinned at the initial 00:00 default.
	cursor, err := unmarshalLoginEventPageToken(&pagination.StreamToken{Cursor: streamState.Cursor}, nil)
	require.NoError(t, err)
	require.Equal(t, "2024-01-01T10:30:00Z", cursor.StartAt)
}

func mustParseDropboxTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(dropbox.TimestampFormat, s)
	require.NoError(t, err)
	return ts
}
