package controllers

import (
	"land/logic"
	"land/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

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
