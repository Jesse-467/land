package controllers

import (
	"land/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CommunityListController(c *gin.Context) {
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("get community list failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, data)
}

func CommunityDetailController(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ResError(c, CodeInvalidParams)
		return
	}

	data, err := logic.GetCommunityDetail(uint64(id))
	if err != nil {
		zap.L().Error("get community detail failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, data)
}
