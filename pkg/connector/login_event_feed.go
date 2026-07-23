package connector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	loginEventFeedID      = "dropbox_login_event_feed"
	loginsEventCategory   = "logins"
	loginSuccessEventType = "login_success"

	// defaultCatchUpWindow bounds how far back the first sync (or a sync
	// recovering from a stale/expired cursor) reaches into the event log.
	defaultCatchUpWindow = 60 * 24 * time.Hour
)

// loginEventFeed emits UsageEvents derived from Dropbox's team event audit log
// (team_log/get_events). Dropbox has no last_login field on team members;
// this is the only way to observe sign-in activity, per Dropbox's own API docs
// and community guidance. Gated behind the sync-user-last-login config flag
// since it requires the events.read scope and can be a high-volume stream on
// active teams.
type loginEventFeed struct {
	client *dropbox.Client
}

func newLoginEventFeed(client *dropbox.Client) *loginEventFeed {
	return &loginEventFeed{client: client}
}

func (f *loginEventFeed) EventFeedMetadata(_ context.Context) *v2.EventFeedMetadata {
	return &v2.EventFeedMetadata{
		Id: loginEventFeedID,
		SupportedEventTypes: []v2.EventType{
			v2.EventType_EVENT_TYPE_USAGE,
		},
	}
}

// loginEventPageToken is the cursor persisted between ListEvents calls. It
// tracks the Dropbox pagination cursor for the current page, plus the sync
// window's start/high-water-mark timestamps for when a new page needs to be
// requested from scratch (no NextPageToken).
type loginEventPageToken struct {
	LatestEventSeen string `json:"latest_event_seen,omitempty"`
	NextPageToken   string `json:"next_page_token,omitempty"`
	StartAt         string `json:"start_at,omitempty"`
}

func unmarshalLoginEventPageToken(token *pagination.StreamToken, defaultStart *timestamppb.Timestamp) (*loginEventPageToken, error) {
	pt := &loginEventPageToken{}
	if token != nil && token.Cursor != "" {
		data, err := base64.StdEncoding.DecodeString(token.Cursor)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, pt); err != nil {
			return nil, err
		}
	}

	if pt.StartAt == "" {
		if defaultStart == nil {
			defaultStart = timestamppb.New(time.Now().Add(-defaultCatchUpWindow))
		}
		pt.StartAt = defaultStart.AsTime().UTC().Format(dropbox.TimestampFormat)
	}
	if pt.LatestEventSeen == "" {
		pt.LatestEventSeen = pt.StartAt
	}

	return pt, nil
}

func (pt *loginEventPageToken) marshal() (string, error) {
	data, err := json.Marshal(pt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (f *loginEventFeed) ListEvents(
	ctx context.Context,
	startAt *timestamppb.Timestamp,
	pToken *pagination.StreamToken,
) ([]*v2.Event, *pagination.StreamState, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	cursor, err := unmarshalLoginEventPageToken(pToken, startAt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("dropbox-connector: failed to unmarshal login event page token: %w", err)
	}

	var payload *dropbox.GetTeamEventsPayload
	var rateLimitData *v2.RateLimitDescription

	if cursor.NextPageToken != "" {
		payload, rateLimitData, err = f.client.GetTeamEventsContinue(ctx, cursor.NextPageToken)
	} else {
		startTime, parseErr := time.Parse(dropbox.TimestampFormat, cursor.StartAt)
		if parseErr != nil {
			l.Debug("dropbox-connector: failed to parse login event start time, using default catch-up window", zap.Error(parseErr))
			startTime = time.Now().Add(-defaultCatchUpWindow)
		}
		payload, rateLimitData, err = f.client.GetTeamEvents(ctx, loginsEventCategory, &startTime, 0)
	}

	var outAnnotations annotations.Annotations
	outAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		return nil, nil, outAnnotations, fmt.Errorf("dropbox-connector: failed to list team login events: %w", err)
	}

	latestEvent, err := time.Parse(dropbox.TimestampFormat, cursor.LatestEventSeen)
	if err != nil {
		latestEvent = time.Unix(0, 0)
	}

	events := make([]*v2.Event, 0, len(payload.Events))
	for _, e := range payload.Events {
		occurredAt, parseErr := time.Parse(dropbox.TimestampFormat, e.Timestamp)
		if parseErr != nil {
			l.Debug("dropbox-connector: skipping login event with unparseable timestamp", zap.String("timestamp", e.Timestamp), zap.Error(parseErr))
			continue
		}
		// Advance the high-water mark from the newest event of ANY type in the
		// logins category, not just successful logins. Otherwise, on a team with
		// no login_success events in the window, LatestEventSeen would never move
		// past its initial now-24h default, so StartAt would never advance and
		// every subsequent sync would re-scan an ever-growing [StartAt, now]
		// window. Every event we see here has already been fully processed, so
		// advancing past it cannot skip a login.
		if occurredAt.After(latestEvent) {
			latestEvent = occurredAt
			cursor.LatestEventSeen = occurredAt.UTC().Format(dropbox.TimestampFormat)
		}

		// The "logins" category also includes logouts, failures, and password
		// resets; only successful sign-ins count as usage.
		if e.EventType.Tag != loginSuccessEventType {
			continue
		}
		// Dropbox logs a team member's own actions under the "admin" actor
		// variant instead of "user" when that member has admin permissions
		// (both wrap the same UserLogInfo shape) — confirmed against a live
		// tenant, where every login_success event came back tagged "admin".
		userInfo := e.Actor.UserInfo()
		if userInfo == nil || userInfo.TeamMemberID == "" {
			continue
		}

		userTrait, traitErr := resourceSdk.NewUserTrait(resourceSdk.WithEmail(userInfo.Email, true))
		if traitErr != nil {
			return nil, nil, outAnnotations, fmt.Errorf("dropbox-connector: failed to build user trait for login event: %w", traitErr)
		}

		events = append(events, &v2.Event{
			// Dropbox's team_log events carry no documented unique event ID, so the
			// actor + timestamp pair is used as a synthetic one (second-granularity
			// timestamps mean same-second repeat logins by one user could collide).
			Id:         fmt.Sprintf("%s-%s", userInfo.TeamMemberID, e.Timestamp),
			OccurredAt: timestamppb.New(occurredAt),
			Event: &v2.Event_UsageEvent{
				UsageEvent: &v2.UsageEvent{
					TargetResource: &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: appResourceType.Id,
							Resource:     dropboxAppResourceID,
						},
						DisplayName: dropboxAppDisplayName,
					},
					ActorResource: &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: userResourceType.Id,
							Resource:     userInfo.TeamMemberID,
						},
						DisplayName: userInfo.DisplayName,
						Annotations: annotations.New(userTrait),
					},
				},
			},
		})
	}

	cursor.NextPageToken = payload.Cursor
	if !payload.HasMore {
		cursor.StartAt = cursor.LatestEventSeen
		cursor.LatestEventSeen = ""
		cursor.NextPageToken = ""
	}

	cursorToken, err := cursor.marshal()
	if err != nil {
		return nil, nil, outAnnotations, fmt.Errorf("dropbox-connector: failed to marshal login event cursor: %w", err)
	}

	return events, &pagination.StreamState{
		Cursor:  cursorToken,
		HasMore: payload.HasMore,
	}, outAnnotations, nil
}
