package dropbox

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const eventsListLimitDefault = 1000

// GetTeamEvents fetches a page of the team event audit log, optionally filtered
// by event category (e.g. "logins") and a start time.
// Based on API: POST /2/team_log/get_events.
func (c *Client) GetTeamEvents(ctx context.Context, category string, startTime *time.Time, limit int) (*GetTeamEventsPayload, *v2.RateLimitDescription, error) {
	if limit == 0 {
		limit = eventsListLimitDefault
	}

	body := GetTeamEventsBody{Limit: limit}
	if category != "" {
		body.Category = &EventCategoryTag{Tag: category}
	}
	if startTime != nil {
		body.Time = &TimeRange{StartTime: startTime.UTC().Format(TimestampFormat)}
	}

	result := &GetTeamEventsPayload{}
	annos, err := c.doRequest(ctx, c.url("/2/team_log/get_events"), http.MethodPost, result, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get team events: %w", err)
	}

	return result, getRateLimitFromAnnos(annos), nil
}

// GetTeamEventsContinue continues a paginated team event audit log listing.
// Based on API: POST /2/team_log/get_events/continue.
func (c *Client) GetTeamEventsContinue(ctx context.Context, cursor string) (*GetTeamEventsPayload, *v2.RateLimitDescription, error) {
	result := &GetTeamEventsPayload{}
	annos, err := c.doRequest(ctx, c.url("/2/team_log/get_events/continue"), http.MethodPost, result, GetTeamEventsContinueBody{Cursor: cursor})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to continue team events: %w", err)
	}

	return result, getRateLimitFromAnnos(annos), nil
}
