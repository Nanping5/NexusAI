package response

import (
	ret "NexusAi/common/code"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: int(ret.CodeSuccess),
		Msg:  ret.CodeSuccess.Msg(),
		Data: data,
	})
}

// SuccessWithMsg 成功响应（自定义消息）
func SuccessWithMsg(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Code: int(ret.CodeSuccess),
		Msg:  msg,
		Data: data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, errCode ret.Code) {
	httpStatus := getHTTPStatus(errCode)
	c.JSON(httpStatus, Response{
		Code: int(errCode),
		Msg:  errCode.Msg(),
	})
}

// FailWithMsg 失败响应（自定义消息）
func FailWithMsg(c *gin.Context, errCode ret.Code, msg string) {
	httpStatus := getHTTPStatus(errCode)
	c.JSON(httpStatus, Response{
		Code: int(errCode),
		Msg:  msg,
	})
}

// FailWithData 失败响应（带数据）
func FailWithData(c *gin.Context, errCode ret.Code, data any) {
	httpStatus := getHTTPStatus(errCode)
	c.JSON(httpStatus, Response{
		Code: int(errCode),
		Msg:  errCode.Msg(),
		Data: data,
	})
}

// getHTTPStatus 根据业务错误码获取 HTTP 状态码
func getHTTPStatus(errCode ret.Code) int {
	switch {
	case errCode == ret.CodeSuccess:
		return http.StatusOK
	case errCode >= 2001 && errCode <= 2010:
		// 客户端错误（参数、认证等）
		return http.StatusBadRequest
	case errCode == ret.CodeUnauthorized:
		return http.StatusUnauthorized
	case errCode == ret.CodeForbidden:
		return http.StatusForbidden
	case errCode == ret.CodeRecordNotFound:
		return http.StatusNotFound
	case errCode >= 4001 && errCode <= 6001:
		// 服务端错误
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
