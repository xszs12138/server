package dto

const (
	LivePlatformBilibili = "bilibili"
)

type LiveBroadcast struct {
	IsLive        bool   `json:"isLive"`
	Platform      string `json:"platform"`
	PlatformLabel string `json:"platformLabel,omitempty"`
	RoomTitle     string `json:"roomTitle"`
	StreamTitle   string `json:"streamTitle"`
	Subtitle      string `json:"subtitle"`
	Cover         string `json:"cover"`
	StreamURL     string `json:"streamUrl"`
}

type LiveBroadcastUpdateRequest struct {
	IsLive      bool   `json:"isLive"`
	Platform    string `json:"platform"`
	RoomTitle   string `json:"roomTitle"`
	StreamTitle string `json:"streamTitle"`
	Subtitle    string `json:"subtitle"`
	Cover       string `json:"cover"`
	StreamURL   string `json:"streamUrl"`
}
