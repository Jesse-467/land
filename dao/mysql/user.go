package mysql

import (
	"land/models"
)

// GetUserById 根据用户ID获取用户信息
// 参数:
//   - uid: 用户ID
//
// 返回值:
//   - user: 用户信息
//   - err: 可能的错误
func GetUserById(id uint64) (user *models.User, err error) {
	user = &models.User{}
	err = db.Where("user_id = ?", id).Find(user).Error
	return user, err
}

// InsertUser 向数据库中插入一条新的用户记录
// 参数:
//   - user: 用户信息
//
// 返回值:
//   - err: 可能的错误
func InsertUser(user *models.User) error {
	err := db.Create(user).Error
	return err
}

// CheckUserExist 检查指定用户名的用户是否存在
// 参数:
//   - username: 用户名
//
// 返回值:
//   - bool: 用户是否存在，存在则返回true，否则返回false
func CheckUserExist(username string) bool {
	var count int64
	db.Model(&models.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// Login 用户登录
// 参数:
//   - user: 用户信息，包含用户名和密码
//
// 返回值:
//   - err: 可能的错误，如用户不存在或密码错误
func Login(user *models.User) error {
	pwd := user.Password
	usr := &models.User{}
	err := db.Where("username = ?", user.Username).First(usr).Error
	if err != nil {
		return err
	}

	if usr.Password != pwd {
		return ErrorInvalidPassword
	}
	return nil
}
