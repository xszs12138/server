package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	categories *service.CategoryService
}

func NewCategoryController(categories *service.CategoryService) *CategoryController {
	return &CategoryController{categories: categories}
}

func (ctl *CategoryController) WebList(c *gin.Context) {
	res, err := ctl.categories.WebList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CategoryController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.categories.AdminList(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		page,
		pageSize,
		c.Query("keyword"),
	)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CategoryController) AdminGetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.categories.AdminGetByID(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		id,
	)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CategoryController) Create(c *gin.Context) {
	var req dto.TaxonomyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.categories.Create(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CategoryController) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.TaxonomyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.categories.Update(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CategoryController) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.categories.Delete(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

type TagController struct {
	tags *service.TagService
}

func NewTagController(tags *service.TagService) *TagController {
	return &TagController{tags: tags}
}

func (ctl *TagController) WebList(c *gin.Context) {
	res, err := ctl.tags.WebList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *TagController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.tags.AdminList(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		page,
		pageSize,
		c.Query("keyword"),
	)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *TagController) AdminGetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.tags.AdminGetByID(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		id,
	)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *TagController) Create(c *gin.Context) {
	var req dto.TaxonomyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.tags.Create(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *TagController) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.TaxonomyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.tags.Update(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *TagController) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.tags.Delete(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		writeTaxonomyError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func writeTaxonomyError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrCategoryNotFound) || errors.Is(err, service.ErrTagNotFound) {
		c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
		return
	}
	if errors.Is(err, service.ErrCategorySlugConflict) || errors.Is(err, service.ErrTagSlugConflict) {
		c.JSON(http.StatusConflict, dto.Error(40900, "资源冲突"))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
}
