package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

// The user resource type is for all user objects from the database.
var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

var groupResourceType = &v2.ResourceType{
	Id:          "group",
	DisplayName: "Group",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}

var roleResourceType = &v2.ResourceType{
	Id:          "role",
	DisplayName: "Role",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
}

// The license resource type models Dropbox team membership types (full vs.
// limited seats). Grants are emitted by userBuilder.Grants from the
// membership_type already fetched during user List(), not from this
// builder, to avoid an O(N) user scan per license type.
//
// membership_type is part of the member profile returned by
// team/members/list_v2, so reading it requires the team_data.member scope.
var licenseResourceType = &v2.ResourceType{
	Id:          "license",
	DisplayName: "License",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_LICENSE_PROFILE},
	Annotations: annotations.New(
		&v2.SkipGrants{},
		&v2.SkipEntitlements{},
		capabilityPermissions("team_data.member"),
	),
}

// capabilityPermissions builds a CapabilityPermissions annotation listing the
// downstream scopes required to sync a resource type.
func capabilityPermissions(permissions ...string) *v2.CapabilityPermissions {
	perms := make([]*v2.CapabilityPermission, 0, len(permissions))
	for _, p := range permissions {
		perms = append(perms, v2.CapabilityPermission_builder{Permission: p}.Build())
	}
	return v2.CapabilityPermissions_builder{Permissions: perms}.Build()
}
