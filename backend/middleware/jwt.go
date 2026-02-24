package middleware

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

type Claims struct {
	UserId uint
	jwt.RegisteredClaims
}

func ReleaseToken(user model.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour * 7)
	claims := &Claims{
		UserId: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vsentry",
			Subject:   "user token",
			Audience:  []string{},
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(viper.GetString("jwt.secret"))
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	fmt.Println(tokenString)
	var token *jwt.Token
	var err error
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		jwtKey := []byte(viper.GetString("jwt.secret"))
		return jwtKey, nil
	})

	return token, claims, err
}
