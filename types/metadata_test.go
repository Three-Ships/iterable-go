package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMetadataResponseRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response any
		body     string
	}{
		{
			name:     "template timestamps and pagination count",
			response: &TemplatesResponse{},
			body:     `{"templates":[{"templateId":25078694,"createdAt":1785461422350,"updatedAt":1789656739064,"name":"Example template","creatorUserId":"creator@example.com","messageTypeId":191047,"clientTemplateId":"example.template"}],"totalTemplatesCount":59,"nextPageUrl":"/api/templates?templateType=Base&messageMedium=Email&sort=id&pageSize=2&page=2"}`,
		},
		{
			name:     "template optional campaign ID",
			response: &TemplatesResponse{},
			body:     `{"templates":[{"templateId":25078694,"createdAt":1785461422350,"updatedAt":1789656739064,"name":"Example template","creatorUserId":"creator@example.com","messageTypeId":191047,"campaignId":19254231}],"totalTemplatesCount":1}`,
		},
		{
			name:     "message type absent optional settings",
			response: &MessageTypeResponse{},
			body:     `{"messageTypes":[{"id":195066,"createdAt":1789495219611,"updatedAt":1789495219611,"name":"Example message type","channelId":147022,"subscriptionPolicy":"OptIn"}]}`,
		},
		{
			name:     "message type explicit zero settings",
			response: &MessageTypeResponse{},
			body:     `{"messageTypes":[{"id":195066,"createdAt":1789495219611,"updatedAt":1789495219611,"name":"Example message type","channelId":147022,"subscriptionPolicy":"OptIn","rateLimitPerMinute":0,"frequencyCap":{"days":0,"messages":0}}]}`,
		},
		{
			name:     "message type signed frequency cap sentinel",
			response: &MessageTypeResponse{},
			body:     `{"messageTypes":[{"id":191047,"createdAt":1785332762577,"updatedAt":1785459808994,"name":"Example message type","channelId":147022,"subscriptionPolicy":"OptIn","frequencyCap":{"days":0,"messages":-1}}]}`,
		},
		{
			name:     "message type positive settings and 64 bit IDs",
			response: &MessageTypeResponse{},
			body:     `{"messageTypes":[{"id":9007199254740993,"createdAt":1789495219611,"updatedAt":1789495219611,"name":"Example message type","channelId":9007199254740995,"subscriptionPolicy":"OptOut","rateLimitPerMinute":10,"frequencyCap":{"days":2,"messages":3}}]}`,
		},
		{
			name:     "campaign explicit empty arrays and pagination count",
			response: &CampaignsResponse{},
			body:     `{"campaigns":[{"id":19197797,"createdAt":1785445662278,"updatedAt":1785458021025,"name":"Example campaign","templateId":25076609,"messageMedium":"Email","createdByUserId":"creator@example.com","updatedByUserId":"editor@example.com","campaignState":"Draft","listIds":[],"suppressionListIds":[],"labels":[],"labelIds":[],"type":"Blast"}],"totalCampaignsCount":3,"nextPageUrl":"/api/campaigns?sort=id&pageSize=2&page=2"}`,
		},
		{
			name:     "campaign absent optional arrays",
			response: &CampaignsResponse{},
			body:     `{"campaigns":[{"id":19197797,"createdAt":1785445662278,"updatedAt":1785458021025,"name":"Example campaign","messageMedium":"Email","createdByUserId":"creator@example.com","campaignState":"Draft","type":"Blast"}]}`,
		},
		{
			name:     "campaign populated arrays",
			response: &CampaignsResponse{},
			body:     `{"campaigns":[{"id":19197797,"createdAt":1785445662278,"updatedAt":1785458021025,"name":"Example campaign","messageMedium":"Email","createdByUserId":"creator@example.com","campaignState":"Ready","listIds":[10675311],"suppressionListIds":[10675312],"labels":["Example"],"labelIds":[123],"type":"Blast"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := json.Unmarshal([]byte(tt.body), tt.response); err != nil {
				t.Fatalf("decode metadata response: %v", err)
			}
			encoded, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("encode metadata response: %v", err)
			}
			got := decodeMetadataJSON(t, string(encoded))
			want := decodeMetadataJSON(t, tt.body)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("metadata round trip = %s, want %s", encoded, tt.body)
			}
		})
	}
}

func TestTemplateRejectsInvalidTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "date string", body: `{"createdAt":"2026-09-01T12:00:00Z"}`},
		{name: "fractional milliseconds", body: `{"createdAt":1785461422350.5}`},
		{name: "overflow", body: `{"updatedAt":9223372036854775808}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var template Template
			if err := json.Unmarshal([]byte(tt.body), &template); err == nil {
				t.Errorf("decode timestamp %s: expected an error", tt.body)
			}
		})
	}
}

func decodeMetadataJSON(t *testing.T, body string) any {
	t.Helper()

	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode JSON for comparison: %v", err)
	}
	return value
}
