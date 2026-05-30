package controller

import (
	"errors"
	"net/http"
	"strings"

	"blog-server/internal/dicttypes"
	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type DictController struct {
	dict *service.DictService
}

func NewDictController(dict *service.DictService) *DictController {
	return &DictController{dict: dict}
}

func (ctl *DictController) AdminListTypes(c *gin.Context) {
	res, err := ctl.dict.AdminListTypes(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *DictController) AdminUpdateType(c *gin.Context) {
	dictType := c.Param("dictType")
	var req dto.DictTypeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}
	if err := ctl.dict.AdminUpdateType(c.Request.Context(), c.GetHeader("Authorization"), dictType, req); err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *DictController) AdminListItems(c *gin.Context) {
	dictType := c.Param("dictType")
	if dictType == "" {
		dictType = c.Query("dictType")
	}

	res, err := ctl.dict.AdminListItems(c.Request.Context(), c.GetHeader("Authorization"), dictType)
	if err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *DictController) AdminCreateItem(c *gin.Context) {
	dictType := c.Param("dictType")
	if dictType == "" {
		dictType = c.Query("dictType")
	}

	var req dto.DictItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}
	if dictType != "" {
		req.DictType = dictType
	}

	res, err := ctl.dict.AdminCreateItem(c.Request.Context(), c.GetHeader("Authorization"), req.DictType, req)
	if err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *DictController) AdminUpdateItem(c *gin.Context) {
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

	res, err := ctl.dict.AdminUpdateItem(c.Request.Context(), c.GetHeader("Authorization"), id, req)
	if err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *DictController) AdminDeleteItem(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	if err := ctl.dict.AdminDeleteItem(c.Request.Context(), c.GetHeader("Authorization"), id); err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *DictController) WebListItems(c *gin.Context) {
	dictType := c.Param("dictType")
	res, err := ctl.dict.WebListItems(c.Request.Context(), dictType)
	if err != nil {
		writeDictError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func writeDictError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return
	}
	if errors.Is(err, service.ErrDictItemNotFound) {
		c.JSON(http.StatusNotFound, dto.Error(40400, "字典项不存在"))
		return
	}
	if errors.Is(err, service.ErrDictItemAlreadyExists) {
		c.JSON(http.StatusConflict, dto.Error(40900, "字典项已存在"))
		return
	}
	if errors.Is(err, dicttypes.ErrUnknownDictType) {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "未知的字典类型"))
		return
	}
	if strings.Contains(err.Error(), "不对博客开放") {
		c.JSON(http.StatusForbidden, dto.Error(40300, err.Error()))
		return
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
}
