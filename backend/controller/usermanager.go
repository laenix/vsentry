package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"golang.org/x/crypto/bcrypt"
)

func ListUser(ctx *gin.Context) {
	var users []model.User
	// 仅查询 ID 和 用户名，不返回密码哈希
	database.GetDB().Select("id", "user_name").Find(&users)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
		"msg":  "success",
	})
}

func AddUser(ctx *gin.Context) {
	DB := database.GetDB()
	// 获取参数
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")

	// 数据验证
	if len(password) < 6 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "密码不能少于6位"})
		return
	}

	if len(name) == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "用户名不能为空"})
		return
	}

	//用户名是否已经存在
	if isUserExist(DB, name) {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "用户名已经存在"})
		return
	}

	// 创建用户
	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": "", "msg": "加密错误"})
		return
	}

	newUser := model.User{
		UserName: name,
		Password: string(hasedPassword),
	}
	DB.Create(&newUser)

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": "", "msg": "注册成功"})
}
func DeleteUser(ctx *gin.Context) {

}
func AdminChangePassword(ctx *gin.Context) {

}
