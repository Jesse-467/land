package controllers

import (
	"fmt"
	"land/models"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enT "github.com/go-playground/validator/v10/translations/en"
	zhT "github.com/go-playground/validator/v10/translations/zh"
)

// 翻译器
var (
	trans ut.Translator
)

// 初始化翻译器
func TransInit(locale string) (err error) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 第一个参数是备用语言环境，后面的参数是支持的环境
		uni := ut.New(en.New(), zh.New(), en.New())

		// 自定义获取tag方法
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// 为SignUpParams结构体注册自定义校验方法
		v.RegisterStructValidation(SignUpParamsValidation, models.SignUpForm{})

		// 翻译
		var ok bool
		trans, ok := uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		switch locale {
		case "zh":
			err = zhT.RegisterDefaultTranslations(v, trans)
		case "en":
			err = enT.RegisterDefaultTranslations(v, trans)
		default:
			err = enT.RegisterDefaultTranslations(v, trans)
		}
		return
	}
	return
}

// 去除提示信息中的结构体名称
func removeStructName(fields map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range fields {
		res[k[strings.Index(k, ".")+1:]] = v
	}
	return res
}

// 自定义SignUpParams结构体校验函数
// 参数sl为结构体级别的校验器
func SignUpParamsValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(models.SignUpForm)

	if s.RePassword != s.Password {
		sl.ReportError(s.RePassword, "re_password", "RePassword", "eqfield", "password")
	}
}
