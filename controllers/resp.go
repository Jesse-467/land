package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RespData struct {
	Code ResCode     `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}

func Res(code ResCode, msg interface{}, data interface{}) *RespData {
	return &RespData{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func ResSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Res(CodeSuccess, CodeSuccess.Msg(), data))
}

func ResError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK, Res(code, code.Msg(), nil))
}

func ResErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(http.StatusOK, Res(code, msg, nil))
}
