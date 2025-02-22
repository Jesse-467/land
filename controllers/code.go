package controllers

type ResCode int64 // 自定义响应状态码

const (
	CodeSuccess           ResCode = 10000 + iota // 成功
	CodeError                                    // 错误
	CodeInvalidParams                            // 无效参数
	CodeNotFound                                 // 未找到
	CodeUnauthorized                             // 未授权
	CodeUserNotFound                             // 用户不存在
	CodeUserExists                               // 用户已存在
	CodeUserPasswordError                        // 用户密码错误
	CodeServerBusy                               // 服务器繁忙

	CodeNeedLogin    // 需要登录
	CodeInvalidToken // 无效的token
)

var (
	CodeMsg = map[ResCode]string{
		CodeSuccess:           "成功",
		CodeError:             "错误",
		CodeInvalidParams:     "无效参数",
		CodeNotFound:          "未找到",
		CodeUnauthorized:      "未授权",
		CodeUserNotFound:      "用户不存在",
		CodeUserExists:        "用户已存在",
		CodeUserPasswordError: "用户密码错误",
		CodeServerBusy:        "服务器繁忙",

		CodeNeedLogin:    "需要登录",
		CodeInvalidToken: "无效的token",
	}
)

func (c ResCode) Msg() string {
	if s, ok := CodeMsg[c]; ok {
		return s
	}
	return CodeMsg[CodeServerBusy]
}
