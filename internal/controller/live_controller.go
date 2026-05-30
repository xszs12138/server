package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type LiveController struct {
	live *service.LiveService
}

func NewLiveController(live *service.LiveService) *LiveController {
	return &LiveController{live: live}
}

func (ctl *LiveController) WebGetLive(c *gin.Context) {
	res, err := ctl.live.WebGetLive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *LiveController) AdminGetLive(c *gin.Context) {
	res, err := ctl.live.AdminGetLive(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeLiveError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *LiveController) AdminUpdateLive(c *gin.Context) {
	var req dto.LiveBroadcastUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.live.AdminUpdateLive(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		writeLiveError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func writeLiveError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrInvalidLivePlatform) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
}
