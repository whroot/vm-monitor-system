package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseCode 统一响应码
type ResponseCode int

const (
	CodeSuccess ResponseCode = 200
	CodeCreated ResponseCode = 201
	CodeBadRequest ResponseCode = 400
	CodeUnauthorized ResponseCode = 401
	CodeForbidden ResponseCode = 403
	CodeNotFound ResponseCode = 404
	CodeConflict ResponseCode = 409
	CodeInternalError ResponseCode = 500
)

// Response 统一响应结构
type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}

// Pagination 分页信息
type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（带自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    CodeCreated,
		Message: "创建成功",
		Data:    data,
	})
}

// Accepted 接受响应（异步任务）
func Accepted(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Code:    202,
		Message: message,
		Data:    data,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: message,
	})
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// Conflict 409错误
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    CodeConflict,
		Message: message,
	})
}

// InternalError 500错误
func InternalError(c *gin.Context, message string, err error) {
	logger.Error("内部错误",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("error", err.Error()),
	)
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: message,
	})
}

// ValidationError 参数验证错误
func ValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    CodeBadRequest,
		Message: "参数错误: " + err.Error(),
	})
}

// PageParam 解析分页参数
func PageParam(c *gin.Context) (page, pageSize int) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	var err error
	page, err = parseInt(pageStr, 1)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = parseInt(pageSizeStr, 20)
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// BuildPagination 构建分页信息
func BuildPagination(page, pageSize, total int) Pagination {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 0 {
		totalPages = 0
	}
	return Pagination{
		Page:      page,
		PageSize:  pageSize,
		Total:     total,
		TotalPages: totalPages,
	}
}

// parseInt 安全解析整数
func parseInt(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultVal, err
	}
	return result, nil
}
