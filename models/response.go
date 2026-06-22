// backend/models/response.go
// 响应处理函数包

package models

import (
	"github.com/gin-gonic/gin"
)

// RequestID 从 Gin Context 中获取请求 ID
func RequestID(c *gin.Context) string {
	if value, ok := c.Get("requestID"); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}

// NewResponse 创建新的 API 响应对象
func NewResponse(c *gin.Context, code int, message string, data interface{}) APIResponse {
	return APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: FormatUTC(UTCNow()),
		RequestID: RequestID(c),
	}
}

// SendStatus 发送状态响应（无数据）
func SendStatus(c *gin.Context, status int, message string) {
	Send(c, status, 0, message, nil)
}

// SendCreated 发送创建成功响应
func SendCreated(c *gin.Context, data interface{}) {
	Send(c, 201, 0, "created", data)
}

// SendNoContent 发送无内容响应
func SendNoContent(c *gin.Context) {
	c.JSON(204, nil)
}

// SendBadRequest 发送错误请求响应
func SendBadRequest(c *gin.Context, message string) {
	Send(c, 400, 1001, message, nil)
}

// SendUnauthorized 发送未授权响应
func SendUnauthorized(c *gin.Context, message string) {
	Send(c, 401, 1002, message, nil)
}

// SendForbidden 发送禁止访问响应
func SendForbidden(c *gin.Context, message string) {
	Send(c, 403, 1003, message, nil)
}

// SendNotFound 发送未找到响应
func SendNotFound(c *gin.Context, message string) {
	Send(c, 404, 2001, message, nil)
}

// SendInternalError 发送内部错误响应
func SendInternalError(c *gin.Context, message string) {
	Send(c, 500, 5000, message, nil)
}
