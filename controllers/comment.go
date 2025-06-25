package controllers

import (
	"fmt"
	"land/dao/mysql"
	"land/models"
	"land/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 创建评论
// @Description 创建评论，需登录
// @Tags 评论相关
// @Accept json
// @Produce json
// @Param data body models.Comment true "评论内容"
// @Success 200 {object} controllers.RespData "创建成功"
// @Failure 400 {object} controllers.RespData "请求参数错误"
// @Router /api/v1/comment [post]
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

// @Summary 评论列表
// @Description 批量获取评论列表
// @Tags 评论相关
// @Accept json
// @Produce json
// @Param ids query []string true "评论ID数组"
// @Success 200 {object} controllers.RespData "评论列表"
// @Failure 400 {object} controllers.RespData "请求参数错误"
// @Router /api/v1/comments [get]
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
