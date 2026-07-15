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

func TestAppBuilder_EntitlementsAndGrants_AreEmpty(t *testing.T) {
	b := newAppBuilder()

	entitlements, _, err := b.Entitlements(context.Background(), nil, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Nil(t, entitlements)

	grants, _, err := b.Grants(context.Background(), nil, resourceSdk.SyncOpAttrs{})
	require.NoError(t, err)
	require.Nil(t, grants)
}
