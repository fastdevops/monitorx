package api

import (
	"github.com/fastdevops/monitorx/global"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginHandler
func LoginHandler(c *gin.Context) {
	session := sessions.Default(c)
	var loginVals struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginVals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 简单的用户名密码验证
	config := global.Config
	if loginVals.Username == config.Auth.Username && loginVals.Password == config.Auth.Password {
		// 登录成功后设置 session
		session.Set("username", loginVals.Username)
		session.Save()

		c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
	}
}

// LogoutHandler - 处理用户退出登录
func LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear() // 清除 session 数据
	session.Save()  // 保存更改
	c.Redirect(http.StatusFound, "/login")
}
