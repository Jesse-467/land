package controllers

import (
	"land/logic"
	"land/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// @Summary 帖子投票
// @Description 对帖子进行投票（赞/踩/取消），需登录
// @Tags 投票相关
// @Accept json
// @Produce json
// @Param data body models.ParamVoteData true "投票参数"
// @Success 200 {object} controllers.RespData "投票成功"
// @Failure 400 {object} controllers.RespData "请求参数错误"
// @Router /api/v1/vote [post]
func PostVoteController(c *gin.Context) {
	p := new(models.ParamVoteData)
	if err := c.ShouldBindJSON(&p); err != nil {
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResError(c, CodeInvalidParams)
			return
		}
		errData := removeStructName(errs.Translate(trans)) // 翻译并去除掉错误提示中的结构体标识
		ResErrorWithMsg(c, CodeInvalidParams, errData)
		return
	}
	// 获取当前请求的用户的id
	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResError(c, CodeNeedLogin)
		return
	}
	// 具体投票的业务逻辑
	if err := logic.VoteForPost(userID, p); err != nil {
		zap.L().Error("logic.VoteForPost() failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, nil)
}
