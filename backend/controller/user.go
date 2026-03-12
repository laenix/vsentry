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

	// GetParameter - := ctx.PostForm("name")
	password := ctx.PostForm("password")

	// DataValidate - len(password) < 6 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "PasswordCannot少于6位"})
		return
	}
	var user model.User
	DB.Where("user_name = ?", name).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "User不存在或PasswordError"})
		return
	}
	// 判断Password是否正确 - err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "User不存在或PasswordError"})
		return
	}

	// 发放token - , err := middleware.ReleaseToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": "", "msg": "System异常"})
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
	// Return结果 - .JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"token": token}, "msg": "LoginSuccess"})
}

func UserChangePassword(ctx *gin.Context) {
	DB := database.GetDB()
	// GetParameter - , _ := ctx.Get("userid")
	oldPassword := ctx.PostForm("old_password")
	newPassword := ctx.PostForm("new_password")
	// DataValidate - len(newPassword) < 6 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "data": "", "msg": "NewPasswordCannot少于6位"})
		return
	}
	var user model.User
	DB.Where("id = ?", userId).First(&user)
	if user.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "User不存在"})
		return
	}
	// 判断旧Password是否正确 - err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "data": "", "msg": "旧PasswordError"})
		return
	}
	// UpdatePassword - , err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "data": "", "msg": "加密Error"})
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
	// Return结果 - .JSON(http.StatusOK, gin.H{"code": 200, "data": "", "msg": "Password修改Success"})
}

func Userinfo(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	u := user.(model.User)
	u.Password = ""
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"user": u}, "msg": "GetUserInfoSuccess"})
}

func isUserExist(db *gorm.DB, name string) bool {
	var user model.User
	db.Where("user_name = ?", name).First(&user)
	if user.ID != 0 {
		return true
	}
	return false
}
