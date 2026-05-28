package controller

import (
	"context"
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	comments *service.CommentService
}

func NewCommentController(comments *service.CommentService) *CommentController {
	return &CommentController{comments: comments}
}

func (ctl *CommentController) WebListByPostSlug(c *gin.Context) {
	res, err := ctl.comments.WebListByPostSlug(c.Request.Context(), c.Param("slug"))
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

func (ctl *CommentController) WebCreate(c *gin.Context) {
	var req dto.WebCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.comments.WebCreate(
		c.Request.Context(),
		c.Param("slug"),
		req,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		if errors.Is(err, service.ErrInvalidCommentParent) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		if errors.Is(err, service.ErrCommentRateLimited) {
			c.JSON(http.StatusTooManyRequests, dto.Error(40000, "评论过于频繁，请稍后再试"))
			return
		}
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CommentController) AdminList(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	postID, err := parseOptionalUint64(c.Query("postId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.comments.AdminList(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		page,
		pageSize,
		postID,
		c.Query("status"),
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

func (ctl *CommentController) AdminListByPostID(c *gin.Context) {
	postID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.comments.AdminListByPostID(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		postID,
		page,
		pageSize,
		c.Query("status"),
	)
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

func (ctl *CommentController) AdminApprove(c *gin.Context) {
	ctl.adminUpdateStatus(c, func(ctx context.Context, authorization string, id uint64) (*dto.CommentStatusResponse, error) {
		return ctl.comments.AdminApprove(ctx, authorization, id)
	})
}

func (ctl *CommentController) AdminReject(c *gin.Context) {
	ctl.adminUpdateStatus(c, func(ctx context.Context, authorization string, id uint64) (*dto.CommentStatusResponse, error) {
		return ctl.comments.AdminReject(ctx, authorization, id)
	})
}

func (ctl *CommentController) AdminDelete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.comments.AdminDelete(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *CommentController) AdminReply(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.AdminCommentReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.comments.AdminReply(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		if errors.Is(err, service.ErrInvalidCommentParent) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *CommentController) adminUpdateStatus(
	c *gin.Context,
	handler func(context.Context, string, uint64) (*dto.CommentStatusResponse, error),
) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := handler(c.Request.Context(), c.GetHeader("Authorization"), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "资源不存在"))
			return
		}
		if errors.Is(err, service.ErrInvalidCommentStatus) {
			c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
