package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func (ctl *AuthController) CreateDictItem(c *gin.Context) {
	if !ctl.ensureAuthenticated(c) {
		return
	}

	var req dto.DictItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.auth.CreateDictItem(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrDictItemAlreadyExists) {
			c.JSON(http.StatusConflict, dto.Error(40900, "字典项已存在"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *AuthController) DeleteDictItem(c *gin.Context) {
	if !ctl.ensureAuthenticated(c) {
		return
	}

	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.auth.DeleteDictItem(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrDictItemNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "字典项不存在"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *AuthController) ListDictItems(c *gin.Context) {
	if !ctl.ensureAuthenticated(c) {
		return
	}

	res, err := ctl.auth.ListDictItems(c.Request.Context(), c.Query("dictType"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *AuthController) UpdateDictItem(c *gin.Context) {
	if !ctl.ensureAuthenticated(c) {
		return
	}

	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	var req dto.DictItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.auth.UpdateDictItem(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrDictItemNotFound) {
			c.JSON(http.StatusNotFound, dto.Error(40400, "字典项不存在"))
			return
		}
		if errors.Is(err, service.ErrDictItemAlreadyExists) {
			c.JSON(http.StatusConflict, dto.Error(40900, "字典项已存在"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
