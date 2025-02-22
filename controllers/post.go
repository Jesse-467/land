package controllers

import (
	"land/logic"
	"land/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetPostListController(c *gin.Context) {
	page, size := GetPageInfo(c)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, data)
	// 返回响应
}

func PostDetailController(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		zap.L().Error("get post detail failed with invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	post, err := logic.GetPostByID(id)
	if err != nil {
		zap.L().Error("logic.GetPostByID(id) failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, post)
}

func CreatePostController(c *gin.Context) {
	p := new(models.Post)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Debug("c.ShouldBindJSON(&p) failed", zap.Error(err))
		zap.L().Error("create post failed", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResError(c, CodeNeedLogin)
		return
	}
	p.AuthorID = userID

	if err = logic.CreatePost(p); err != nil {
		zap.L().Error("logic.CreatePost(p) failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, nil)
}

// GetPostListHandler2 升级版帖子列表接口
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
func GetPostListHandler2(c *gin.Context) {
	p := &models.ParamPostList{
		Page:  1,
		Size:  10,
		Order: models.OrderTime,
	}

	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetPostListHandler2 with invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}
	data, err := logic.GetPostListNew(p) // 更新：合二为一
	// 获取数据
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResSuccess(c, data)

}
