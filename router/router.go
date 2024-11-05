package router

import (
	"github.com/fastdevops/monitorx/api"
	"github.com/fastdevops/monitorx/logger"
	"github.com/fastdevops/monitorx/utils"
	"net/http"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitRouterConfig(cf *utils.Config) {
	r := gin.New()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// limit
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		panic(err)
	}

	// static load
	// r.LoadHTMLFiles("template/login.html")
	r.LoadHTMLGlob("template/*")
	r.Static("/static", "./static")

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// 设置 session 存储引擎，这里使用 cookie 存储
	store := cookie.NewStore([]byte(cf.Session.Secret))
	r.Use(sessions.Sessions("mysession", store))
	r.POST("/login", api.LoginHandler)

	// 登录后的监控页面接口，受保护，只有登录后才可访问
	authorized := r.Group("/")
	authorized.Use(utils.AuthRequired())
	{
		authorized.GET("/dashboard/hadoop", api.DashboardHandler)
		authorized.GET("/logout", api.LogoutHandler)
	}

	// run
	go func() {
		if err := r.Run(":21000"); err == nil {
			logger.Logger.Fatal("Service failed to start", zap.Error(err))
		}
	}()

	// start log
	logger.Logger.Info("Service started successfully on port 21000")

	select {}
}
