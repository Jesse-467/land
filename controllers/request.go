package controllers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "userID"
)

var (
	ErrorUserNotLogin = errors.New("User not login")
)

// GetCurrentUserID 获取当前登录的用户ID
// 参数:
//   - c: gin的上下文
//
// 返回值:
//   - userID: 用户ID
//   - err: 可能的错误，如用户未登录
func GetCurrentUserID(c *gin.Context) (userID uint64, err error) {
	uID, ok := c.Get(ContextUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uID.(uint64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

// GetPageInfo 从请求中获取分页信息
// 参数:
//   - c: gin的上下文
//
// 返回值:
//   - page: 页码，默认为1
//   - size: 每页大小，默认为50，最大为100
func GetPageInfo(c *gin.Context) (int64, int64) {
	PageStr := c.Query("page")
	SizeStr := c.Query("size")
	var (
		page int64
		size int64
		err  error
	)
	page, err = strconv.ParseInt(PageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1 // 如果解析失败或小于1，默认为第1页
	}
	size, err = strconv.ParseInt(SizeStr, 10, 64)
	if err != nil || size < 1 {
		size = 50 // 如果解析失败或小于1，默认每页50条
	}
	if size > 100 {
		size = 100 // 限制每页最多100条
	}
	return page, size
}
