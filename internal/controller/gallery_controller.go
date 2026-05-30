package controller

import (
	"net/http"
	"strconv"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type GalleryController struct {
	gallery *service.GalleryService
}

func NewGalleryController(gallery *service.GalleryService) *GalleryController {
	return &GalleryController{gallery: gallery}
}

func parseOptionalIntQuery(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func (ctl *GalleryController) WebListImages(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	albumID, err := parseOptionalIntQuery(c.Query("albumId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.gallery.WebListImages(
		c.Request.Context(),
		page,
		pageSize,
		albumID,
		c.Query("order"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
