package logic

import (
	"crypto/md5"
	"encoding/hex"
	"land/dao/mysql"
	"land/models"
	"land/pkg/jwt"
	"land/pkg/snowflake"
)

const (
	secret = "0d000721"
)

// SignUp 处理用户注册逻辑
// 参数:
//   - p: 用户注册参数
//
// 返回值:
//   - error: 可能发生的错误
func SignUp(p *models.SignUpForm) error {
	// 是否该Username已被注册

	if ok := mysql.CheckUserExist(p.UserName); ok {
		return mysql.ErrorUserExist
	}

	id := snowflake.GetID()

	// 注册用户
	user := models.User{
		UserID:   id,
		Username: p.UserName,
		Password: encryptPassword(p.Password),
		Email:    p.Email,
	}

	return mysql.InsertUser(&user)
}

// Login 处理用户登录逻辑
// 参数:
//   - p: 用户登录参数
//
// 返回值:
//   - *models.User: 登录成功的用户信息
//   - error: 可能发生的错误
func Login(p *models.LoginForm) (user *models.User, err error) {
	if ok := mysql.CheckUserExist(p.UserName); !ok {
		return nil, mysql.ErrorUserNotExist
	}

	user = &models.User{
		Username: p.UserName,
		Password: encryptPassword(p.Password),
	}

	// 验证用户登录信息
	if err := mysql.Login(user); err != nil {
		return nil, err
	}
	// 生成JWT令牌
	token, err := jwt.GenToken(int64(user.UserID), user.Username)
	if err == nil {
		user.Token = token
	}
	return
}

// encryptPassword 密码加密
// 参数:
//   - oPassword: 原始密码
//
// 返回值:
//   - 加密后的密码
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}
