package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCampaign_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 12345,
		"name": "Spring Sale",
		"createdAt": 1700000000,
		"updatedAt": 1700001000,
		"campaignState": "Finished",
		"type": "Blast",
		"labels": ["promotions", "spring"],
		"labelIds": [101, 102]
	}`

	var c Campaign
	err := json.Unmarshal([]byte(raw), &c)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), c.Id)
	assert.Equal(t, "Spring Sale", c.Name)
	assert.Equal(t, []string{"promotions", "spring"}, c.Labels)
	assert.Equal(t, []int64{101, 102}, c.LabelIds)
}

func TestCampaignsResponse_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"campaigns": [
			{"id": 1, "name": "First"}
		],
		"nextPageUrl": "/api/campaigns?page=2&pageSize=1000&sort=id",
		"previousPageUrl": "/api/campaigns?page=1&pageSize=1000&sort=id",
		"totalCampaignsCount": 42
	}`

	var res CampaignsResponse
	err := json.Unmarshal([]byte(raw), &res)
	require.NoError(t, err)

	assert.Len(t, res.Campaigns, 1)
	assert.Equal(t, "/api/campaigns?page=2&pageSize=1000&sort=id", res.NextPageUrl)
	assert.Equal(t, "/api/campaigns?page=1&pageSize=1000&sort=id", res.PreviousPageUrl)
	assert.Equal(t, int64(42), res.TotalCampaignsCount)
}
