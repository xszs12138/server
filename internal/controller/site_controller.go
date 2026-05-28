package controller

import (
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
