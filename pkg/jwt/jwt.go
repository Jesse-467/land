package jwt

import (
	"errors"
	"land/settings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

// TokenExpiredDuration 定义了token的过期时间，设置为30天
const TokenExpiredDuration = time.Hour * 24 * 30

// mySecret 是用于签名的密钥
// var mySecret = []byte("0d000721")

// MyClaims 自定义声明结构体并内嵌jwt.StandardClaims
// jwt包自带的jwt.StandardClaims只包含了官方字段
// 我们这里需要额外记录UserID和Username字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenToken 生成JWT
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//
// 返回：
//   - string: 生成的token字符串
//   - error: 可能发生的错误
func GenToken(userID int64, username string) (string, error) {
	// 创建一个我们自己的声明数据
	c := MyClaims{
		userID,
		username, // 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(
				time.Duration(viper.GetInt("auth.jwt_expire")) * time.Hour).Unix(), // 过期时间
			Issuer: "jesse", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString([]byte(settings.Conf.Secret))
}

// ParseToken 解析JWT
// 参数：
//   - tokenString: 待解析的token字符串
//
// 返回：
//   - *MyClaims: 解析后的声明结构体指针
//   - error: 可能发生的错误
func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析token
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(settings.Conf.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid { // 校验token
		return mc, nil
	}
	return nil, errors.New("无效的token")
}
