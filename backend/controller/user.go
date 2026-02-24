package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/middleware"
	"github.com/laenix/vsentry/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Login(ctx *gin.Context) {
	DB := database.GetDB()

	// 获取参数
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")

	// 数据验证
	if len(password) < 6 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "密码不能少于6位"})
		return
	}
	var user model.User
	DB.Where("user_name = ?", name).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "用户不存在或密码错误"})
		return
	}
	// 判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "用户不存在或密码错误"})
		return
	}

	// 发放token
	token, err := middleware.ReleaseToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": "", "msg": "系统异常"})
		log.Printf("token generate error : %v", err)
		return
	}
	//
	loginlogs := model.UserLoginLogs{
		UserID:    user.ID,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}
	DB.Create(&loginlogs)
	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"token": token}, "msg": "登录成功"})
}

func UserChangePassword(ctx *gin.Context) {
	DB := database.GetDB()
	// 获取参数
	userId, _ := ctx.Get("userid")
	oldPassword := ctx.PostForm("old_password")
	newPassword := ctx.PostForm("new_password")
	// 数据验证
	if len(newPassword) < 6 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "新密码不能少于6位"})
		return
	}
	var user model.User
	DB.Where("id = ?", userId).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "用户不存在"})
		return
	}
	// 判断旧密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "旧密码错误"})
		return
	}
	// 更新密码
	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": "", "msg": "加密错误"})
		return
	}
	DB.Model(&user).Update("password", string(hasedPassword))
	//
	actionlogs := model.UserActionLogs{
		UserID:    user.ID,
		Action:    "change password",
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}
	DB.Create(&actionlogs)
	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": "", "msg": "密码修改成功"})
}

func Userinfo(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	u := user.(model.User)
	u.Password = ""
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": u}, "msg": "获取用户信息成功"})
}

func isUserExist(db *gorm.DB, name string) bool {
	var user model.User
	db.Where("user_name = ?", name).First(&user)
	if user.ID != 0 {
		return true
	}
	return false
}
