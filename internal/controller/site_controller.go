package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type SiteController struct {
	sites *service.SiteService
}

func NewSiteController(sites *service.SiteService) *SiteController {
	return &SiteController{sites: sites}
}

func (ctl *SiteController) WebGetSite(c *gin.Context) {
	res, err := ctl.sites.WebGetSite(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *SiteController) AdminGetSiteSettings(c *gin.Context) {
	res, err := ctl.sites.AdminGetSiteSettings(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeSiteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *SiteController) AdminUpdateSiteSettings(c *gin.Context) {
	var req dto.AdminSiteSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.sites.AdminUpdateSiteSettings(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		writeSiteError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func writeSiteError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrInvalidSiteSettings) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "站点设置参数无效"))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
}
