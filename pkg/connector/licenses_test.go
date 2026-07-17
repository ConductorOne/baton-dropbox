package connector

import (
	"context"
	"testing"

	"github.com/conductorone/baton-dropbox/pkg/connector/dropbox"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/require"
)

func TestLicenseResource_Trait(t *testing.T) {
	res, err := licenseResource(fullLicenseType)
	require.NoError(t, err)

	profile, err := resourceSdk.GetLicenseProfileTrait(res)
	require.NoError(t, err, "license resource must carry a LicenseProfileTrait")

	require.Equal(t, "full", profile.GetLicenseName())
	require.Equal(t, []string{"license:full:assigned"}, profile.GetEntitlementIds())
}

func TestUserBuilder_Grants_FullMembershipGrantsLicense(t *testing.T) {
	res, err := userResource(dropbox.Profile{
		AccountID:      "acc-1",
		TeamMemberID:   "dbmid:1",
		Email:          "user@example.com",
		Status:         dropbox.Tag{Tag: "active"},
		MembershipType: dropbox.Tag{Tag: "full"},
	}, nil)
	require.NoError(t, err)

	o := &userBuilder{}
	grants, _, err := o.Grants(context.Background(), res, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Len(t, grants, 1)
	require.Equal(t, "license:full:assigned", grants[0].Entitlement.Id)
	require.Equal(t, "license", grants[0].Entitlement.Resource.Id.ResourceType)
	require.Equal(t, "full", grants[0].Entitlement.Resource.Id.Resource)
	require.Equal(t, res.Id.Resource, grants[0].Principal.Id.Resource)
}

func TestUserBuilder_Grants_LimitedMembershipNoLicense(t *testing.T) {
	res, err := userResource(dropbox.Profile{
		AccountID:      "acc-2",
		TeamMemberID:   "dbmid:2",
		Email:          "limited@example.com",
		Status:         dropbox.Tag{Tag: "active"},
		MembershipType: dropbox.Tag{Tag: "limited"},
	}, nil)
	require.NoError(t, err)

	o := &userBuilder{}
	grants, _, err := o.Grants(context.Background(), res, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Empty(t, grants)
}

// TestUserBuilder_Grants_SeatStatusMatrix verifies that a license grant is
// emitted only for full members whose status consumes a seat (active,
// suspended) and never for invited/removed members — regardless of
// membership_type — so departed and not-yet-joined members don't over-count
// consumed seats. See CXP-763 / licenseSeatStatuses.
func TestUserBuilder_Grants_SeatStatusMatrix(t *testing.T) {
	cases := []struct {
		status         string
		membershipType string
		wantGrant      bool
	}{
		{"active", "full", true},
		{"suspended", "full", true},
		{"invited", "full", false},
		{"removed", "full", false},
		{"active", "limited", false},
		{"suspended", "limited", false},
	}

	o := &userBuilder{}
	for _, tc := range cases {
		t.Run(tc.status+"_"+tc.membershipType, func(t *testing.T) {
			res, err := userResource(dropbox.Profile{
				AccountID:      "acc",
				TeamMemberID:   "dbmid:x",
				Email:          "user@example.com",
				Status:         dropbox.Tag{Tag: tc.status},
				MembershipType: dropbox.Tag{Tag: tc.membershipType},
			}, nil)
			require.NoError(t, err)

			grants, _, err := o.Grants(context.Background(), res, resourceSdk.SyncOpAttrs{})
			require.NoError(t, err)
			if tc.wantGrant {
				require.Len(t, grants, 1)
				require.Equal(t, "license:full:assigned", grants[0].Entitlement.Id)
			} else {
				require.Empty(t, grants)
			}
		})
	}
}
