package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type MusicController struct {
	music *service.MusicService
}

func NewMusicController(music *service.MusicService) *MusicController {
	return &MusicController{music: music}
}

func (ctl *MusicController) WebPlaylist(c *gin.Context) {
	res, err := ctl.music.WebPlaylist(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *MusicController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var visible *bool
	switch c.Query("visible") {
	case "true":
		v := true
		visible = &v
	case "false":
		v := false
		visible = &v
	}

	res, err := ctl.music.AdminList(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		page,
		pageSize,
		c.Query("keyword"),
		visible,
	)
	if err != nil {
		writeMusicError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *MusicController) AdminGetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.music.AdminGetByID(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		id,
	)
	if err != nil {
		writeMusicError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *MusicController) Create(c *gin.Context) {
	var req dto.MusicTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.music.Create(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		writeMusicError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(map[string]uint64{"id": res.ID}))
}

func (ctl *MusicController) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.MusicTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.music.Update(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		writeMusicError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(map[string]uint64{"id": res.ID}))
}

func (ctl *MusicController) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.music.Delete(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		writeMusicError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func writeMusicError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrMusicTrackNotFound) {
		c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
}
