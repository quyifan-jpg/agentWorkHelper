package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userId uint, secret string, expire int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Duration(expire) * time.Second).Unix()
	claims["iat"] = time.Now().Unix()
	claims["uid"] = userId

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["uid"].(float64); ok {
			return uint(uid), nil
		}
	}

	return 0, jwt.ErrTokenInvalidClaims
}
