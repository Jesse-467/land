package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

const (
	TokenExpiredDuration = time.Hour * 24 * 30 // 30天过期时间
)

var (
	secretkey = []byte("0d000721") // 密钥
)

/*
自定义Claims结构体
包含需要传输的用户信息和jwt标准字段
*/
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	*jwt.StandardClaims
}

func GenerateToken(userID int64, username string) (string, error) {
	c := JWTClaims{
		UserID:   userID,
		Username: username,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(
				time.Duration(viper.GetInt("auth.jwt_expire")) * time.Second).Unix(),
			Issuer: "land",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(secretkey)
}

func ParseToken(tokenString string) (*JWTClaims, error) {
	c := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, c, func(token *jwt.Token) (interface{}, error) {
		return secretkey, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return c, nil
	}
	return nil, errors.New("token is invalid")
}
