package controllers

import (
	"fmt"
	"land/dao/mysql"
	"land/models"
	"land/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommentHandler 创建评论
func CommentHandler(c *gin.Context) {
	var comment models.Comment
	if err := c.BindJSON(&comment); err != nil {
		fmt.Println(err)
		ResError(c, CodeInvalidParams)
		return
	}

	// 生成评论ID
	commentID := snowflake.GetID()

	// 获取作者ID，当前请求的UserID
	userID, err := GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("GetCurrentUserID() failed", zap.Error(err))
		ResError(c, CodeNeedLogin)
		return
	}
	comment.CommentID = commentID
	comment.AuthorID = userID

	// 创建评论
	if err := mysql.CreateComment(&comment); err != nil {
		zap.L().Error("mysql.CreateComment(&comment) failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, nil)
}

// CommentListHandler 评论列表
func CommentListHandler(c *gin.Context) {
	ids, ok := c.GetQueryArray("ids")
	if !ok {
		ResError(c, CodeInvalidParams)
		return
	}
	posts, err := mysql.GetCommentListByIDs(ids)
	if err != nil {
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, posts)
}
