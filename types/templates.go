package types

// Template contains project template metadata with epoch-millisecond timestamps.
type Template struct {
	TemplateId       int64  `json:"templateId"`
	CampaignId       int64  `json:"campaignId,omitempty"`
	ClientTemplateId string `json:"clientTemplateId,omitempty"`
	CreatedAt        int64  `json:"createdAt"`
	CreatorUserId    string `json:"creatorUserId"`
	MessageTypeId    int64  `json:"messageTypeId"`
	Name             string `json:"name"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// TemplatesResponse contains one page of project template metadata.
type TemplatesResponse struct {
	Templates           []Template `json:"templates"`
	NextPageUrl         string     `json:"nextPageUrl,omitempty"`
	PreviousPageUrl     string     `json:"previousPageUrl,omitempty"`
	TotalTemplatesCount int64      `json:"totalTemplatesCount,omitempty"`
}
