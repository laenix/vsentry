package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer") {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401, "msg": "Token验证失败",
			})
			ctx.Abort()
			return
		}
		tokenString = tokenString[7:]
		token, Claims, err := ParseToken(tokenString)
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401, "msg": "Token失效，请重新登录", "data": err,
			})
			ctx.Abort()
			return
		}
		userId := Claims.UserId
		DB := database.GetDB()
		var user model.User
		DB.Table("users").Where("id = ?", userId).Scan(&user)
		if user.ID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401, "msg": "用户不存在",
			})
			ctx.Abort()
			return
		}
		ctx.Set("user", user)
		ctx.Set("userid", user.ID)
		ctx.Next()
	}
}
