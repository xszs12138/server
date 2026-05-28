package controller

import (
	"errors"
	"net/http"
	"strconv"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	posts *service.PostService
}

func NewPostController(posts *service.PostService) *PostController {
	return &PostController{posts: posts}
}

func parseOptionalUint64(value string) (uint64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

func (ctl *PostController) WebList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.WebList(
		c.Request.Context(),
		page,
		pageSize,
		c.Query("categorySlug"),
		c.Query("tagSlug"),
		c.Query("keyword"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) WebArchives(c *gin.Context) {
	res, err := ctl.posts.WebArchives(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) WebGetBySlug(c *gin.Context) {
	res, err := ctl.posts.WebGetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	categoryID, err := parseOptionalUint64(c.Query("categoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}
	tagID, err := parseOptionalUint64(c.Query("tagId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.AdminList(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		page,
		pageSize,
		c.Query("status"),
		categoryID,
		tagID,
		c.Query("keyword"),
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) AdminGetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.AdminGetByID(c.Request.Context(), c.GetHeader("Authorization"), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) Create(c *gin.Context) {
	var req dto.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.Create(c.Request.Context(), c.GetHeader("Authorization"), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrInvalidPostStatus) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		if errors.Is(err, service.ErrCategoryNotFound) || errors.Is(err, service.ErrTagNotFound) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		if errors.Is(err, service.ErrPostSlugConflict) {
			c.JSON(http.StatusConflict, dto.Error(40900, "资源冲突"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.Update(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		if errors.Is(err, service.ErrInvalidPostStatus) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		if errors.Is(err, service.ErrCategoryNotFound) || errors.Is(err, service.ErrTagNotFound) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		if errors.Is(err, service.ErrPostSlugConflict) {
			c.JSON(http.StatusConflict, dto.Error(40900, "资源冲突"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *PostController) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.posts.Delete(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *PostController) UpdateStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.PostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.posts.UpdateStatus(c.Request.Context(), c.GetHeader("Authorization"), id, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		if errors.Is(err, service.ErrInvalidPostStatus) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
