package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

// licenseAssigned is the entitlement slug that grants a Dropbox team member
// a "full" license seat.
const licenseAssigned = "assigned"

// fullLicenseType is the only Dropbox team_member membership_type that
// consumes a license seat. "limited" members (deprecated by Dropbox, but
// still returned for legacy tenants) do not use the shared quota and are
// intentionally not modeled as a License resource, so no License resource
// is synced for them and no grant is emitted by userBuilder.Grants.
const fullLicenseType = "full"

type licenseBuilder struct{}

func (b *licenseBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return licenseResourceType
}

// List returns the single static "full" license resource. Dropbox has no
// /licenses endpoint; membership_type is a fixed enum already returned on
// every team/members/list_v2 response.
func (b *licenseBuilder) List(_ context.Context, _ *v2.ResourceId, _ resourceSdk.SyncOpAttrs) ([]*v2.Resource, *resourceSdk.SyncOpResults, error) {
	res, err := licenseResource(fullLicenseType)
	if err != nil {
		return nil, nil, err
	}
	return []*v2.Resource{res}, &resourceSdk.SyncOpResults{}, nil
}

// StaticEntitlements returns a single "assigned" entitlement shared across
// all license resources. Grants are emitted by userBuilder.Grants.
func (b *licenseBuilder) StaticEntitlements(_ context.Context, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return []*v2.Entitlement{
		entitlement.NewAssignmentEntitlement(
			nil,
			licenseAssigned,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName("Assigned"),
			entitlement.WithDescription("User holds a full Dropbox license seat"),
		),
	}, nil, nil
}

func (b *licenseBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Entitlement, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants is never called: licenseResourceType carries SkipGrants, so license
// grants are emitted by userBuilder.Grants from the membership_type already
// fetched during user List().
func (b *licenseBuilder) Grants(_ context.Context, _ *v2.Resource, _ resourceSdk.SyncOpAttrs) ([]*v2.Grant, *resourceSdk.SyncOpResults, error) {
	return nil, nil, nil
}

func newLicenseBuilder() *licenseBuilder {
	return &licenseBuilder{}
}

// licenseResource builds the License resource for a Dropbox membership_type.
// The membership type name is used as the stable resource ID because
// Dropbox's TeamMembershipType enum is fixed.
func licenseResource(membershipType string) (*v2.Resource, error) {
	licenseEntitlementID := entitlement.NewEntitlementID(
		&v2.Resource{Id: &v2.ResourceId{ResourceType: licenseResourceType.Id, Resource: membershipType}},
		licenseAssigned,
	)

	return resourceSdk.NewResource(
		membershipType,
		licenseResourceType,
		membershipType,
		resourceSdk.WithLicenseProfileTrait(
			resourceSdk.WithLicenseName(membershipType),
			resourceSdk.WithLicenseEntitlementIDs(licenseEntitlementID),
		),
	)
}
