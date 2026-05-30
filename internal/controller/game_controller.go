package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	games *service.GameService
}

func NewGameController(games *service.GameService) *GameController {
	return &GameController{games: games}
}

func (ctl *GameController) WebList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.games.WebList(
		c.Request.Context(),
		page,
		pageSize,
		c.Query("genre"),
		c.Query("status"),
		c.Query("sort"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) WebListGenres(c *gin.Context) {
	res, err := ctl.games.WebListGenres(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) WebSidebar(c *gin.Context) {
	res, err := ctl.games.WebSidebar(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.games.AdminList(c.Request.Context(), c.GetHeader("Authorization"), page, pageSize)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) AdminGetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.games.AdminGetByID(c.Request.Context(), c.GetHeader("Authorization"), id)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) AdminUpdate(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.AdminGameUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.games.AdminUpdate(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *GameController) AdminSync(c *gin.Context) {
	res, err := ctl.games.AdminSync(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeGameError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func writeGameError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrGameNotFound) {
		c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
		return
	}
	if errors.Is(err, service.ErrSteamNotConfigured) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "Steam 未配置：请检查 config/config.yaml 的 steam.apiKey 与 steam.steamId"))
		return
	}
	if errors.Is(err, service.ErrSteamLibraryPrivate) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, err.Error()))
		return
	}
	if errors.Is(err, service.ErrSteamInvalidAPIKey) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, err.Error()))
		return
	}
	if errors.Is(err, service.ErrInvalidPlayStatus) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, err.Error()))
}
