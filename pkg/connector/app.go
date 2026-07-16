package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	entitlementSdk "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const (
	dropboxAppResourceID  = "dropbox"
	dropboxAppDisplayName = "Dropbox"

	// appAccessEntitlement is the slug of the "access" entitlement on the
	// Dropbox app resource. loginEventFeed's UsageEvents target this resource,
	// and C1's usage uplift reads those usage principals through this
	// entitlement (see baton-okta / baton-aws for the same pattern).
	appAccessEntitlement = "access"
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

// StaticEntitlements returns a single "access" assignment entitlement shared by
// the (singleton) Dropbox app resource. C1's usage uplift
// (uplift_entitlement_usage_v2) iterates an app's App-trait entitlements and
// reads usage principals keyed to each entitlement's resource, so
// loginEventFeed's UsageEvents only surface if this entitlement exists. Grants
// are intentionally not emitted: the usage uplift maps the login actor directly
// to a synced app user, not via a grant.
func (b *appBuilder) StaticEntitlements(_ context.Context, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return []*v2.Entitlement{
		entitlementSdk.NewAssignmentEntitlement(
			nil,
			appAccessEntitlement,
			entitlementSdk.WithGrantableTo(userResourceType),
			entitlementSdk.WithDisplayName("Dropbox Access"),
			entitlementSdk.WithDescription("Has access to Dropbox"),
		),
	}, nil, nil
}

func (b *appBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func (b *appBuilder) Grants(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}
