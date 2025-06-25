package controllers

import (
	"land/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 社区列表
// @Description 获取所有社区列表
// @Tags 社区相关
// @Accept json
// @Produce json
// @Success 200 {object} controllers.RespData "社区列表"
// @Failure 400 {object} controllers.RespData "请求参数错误"
// @Router /api/v1/community [get]
func CommunityListController(c *gin.Context) {
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("get community list failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, data)
}

// @Summary 社区详情
// @Description 获取指定社区的详细信息
// @Tags 社区相关
// @Accept json
// @Produce json
// @Param id path int true "社区ID"
// @Success 200 {object} controllers.RespData "社区详情"
// @Failure 400 {object} controllers.RespData "请求参数错误"
// @Router /api/v1/community/{id} [get]
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
