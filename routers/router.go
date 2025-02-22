package routers

import (
	"land/controllers"
	"land/logger"
	"land/middlewares"

	"github.com/gin-gonic/gin"
)

func SetRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	v1 := r.Group("/api/v1")
	// 登陆注册相关
	v1.POST("/login", controllers.LoginHandler)     // 登录
	v1.POST("/register", controllers.SignUpHandler) // 注册

	// 为后续路由启用JWT验证中间件

	v1.Use(middlewares.JWTAuth())

	{
		// 社区相关
		v1.GET("/community", controllers.CommunityListController)       // 获取社区列表
		v1.GET("/community/:id", controllers.CommunityDetailController) // 获取社区详情

		// 帖子相关
		v1.GET("/post", controllers.GetPostListController)    // 获取帖子列表
		v1.GET("/post/:id", controllers.PostDetailController) // 获取帖子详情
		v1.POST("/post", controllers.CreatePostController)    // 创建帖子
		v1.POST("/vote", controllers.PostVoteController)      // 帖子投票
		v1.GET("/posts2/", controllers.GetPostListHandler2)   // 根据时间或分数获取帖子列表（优化版）
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code": 404,
			"msg":  "Not Found",
		})
	})

	return r
}
