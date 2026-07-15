package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const (
	dropboxAppResourceID  = "dropbox"
	dropboxAppDisplayName = "Dropbox"
)

// appBuilder syncs a single, static "Dropbox" App resource. Dropbox has no
// concept of multiple discrete apps the way an IdP connector would; this
// resource exists only so that loginEventFeed has a synced TargetResource to
// attach last-login usage events to.
type appBuilder struct{}

func newAppBuilder() *appBuilder {
	return &appBuilder{}
}

func (b *appBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return appResourceType
}

func (b *appBuilder) List(_ context.Context, _ *v2.ResourceId, _ resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	res, err := resourceSdk.NewAppResource(dropboxAppDisplayName, appResourceType, dropboxAppResourceID, nil)
	if err != nil {
		return nil, nil, err
	}
	return []*v2.Resource{res}, &resourceSdk.SyncOpResults{}, nil
}

func (b *appBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func (b *appBuilder) Grants(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}
