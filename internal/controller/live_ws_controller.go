package controller

import (
	"encoding/json"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/livehub"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var liveWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LiveWSController struct {
	live *service.LiveService
	hub  *livehub.Hub
}

func NewLiveWSController(live *service.LiveService, hub *livehub.Hub) *LiveWSController {
	return &LiveWSController{live: live, hub: hub}
}

// WebLiveWS GET /api/web/live/ws — 订阅直播状态推送
func (ctl *LiveWSController) WebLiveWS(c *gin.Context) {
	conn, err := liveWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ctl.hub.Register(conn)
	defer client.Unregister()

	snapshot, err := ctl.live.WebGetLive(c.Request.Context())
	if err == nil && snapshot != nil {
		payload, marshalErr := json.Marshal(dto.LiveWSMessage{
			Type: dto.LiveWSTypeUpdated,
			Data: snapshot,
		})
		if marshalErr == nil {
			client.Enqueue(payload)
		}
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
