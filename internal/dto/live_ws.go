package dto

const LiveWSTypeUpdated = "live.updated"

// LiveWSMessage WebSocket 下行消息（博客端订阅直播状态）
type LiveWSMessage struct {
	Type string         `json:"type"`
	Data *LiveBroadcast `json:"data"`
}
