package connector

import (
	"context"
	"testing"

	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/require"
)

func TestAppBuilder_List_ReturnsSingleStaticDropboxApp(t *testing.T) {
	b := newAppBuilder()

	resources, _, err := b.List(context.Background(), nil, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	res := resources[0]
	require.Equal(t, appResourceType.Id, res.Id.ResourceType)
	require.Equal(t, dropboxAppResourceID, res.Id.Resource)
	require.Equal(t, dropboxAppDisplayName, res.DisplayName)
}

func TestAppBuilder_Entitlements_ReturnsAccessEntitlement(t *testing.T) {
	b := newAppBuilder()

	res, _, err := b.List(context.Background(), nil, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)

	entitlements, _, err := b.Entitlements(context.Background(), res[0], resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Len(t, entitlements, 1)
	require.Equal(t, appAccessEntitlement, entitlements[0].Slug)
	require.Equal(t, res[0].Id.Resource, entitlements[0].Resource.Id.Resource)
}

func TestAppBuilder_Grants_AreEmpty(t *testing.T) {
	b := newAppBuilder()

	grants, _, err := b.Grants(context.Background(), nil, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Nil(t, grants)
}
