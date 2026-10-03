package types

// MessageType contains subscription settings and epoch-millisecond timestamps.
type MessageType struct {
	Id                 int64         `json:"id"`
	CreatedAt          int64         `json:"createdAt"`
	UpdatedAt          int64         `json:"updatedAt"`
	Name               string        `json:"name"`
	ChannelId          int64         `json:"channelId"`
	SubscriptionPolicy string        `json:"subscriptionPolicy"`
	RateLimitPerMinute *int64        `json:"rateLimitPerMinute,omitempty"`
	FrequencyCap       *FrequencyCap `json:"frequencyCap,omitempty"`
}

// FrequencyCap contains a message limit over a period of days.
// Messages is signed to preserve API sentinel values such as -1.
type FrequencyCap struct {
	Days     int `json:"days"`
	Messages int `json:"messages"`
}

// MessageTypeResponse contains the project's message types.
type MessageTypeResponse struct {
	MessageTypes []MessageType `json:"messageTypes"`
}
