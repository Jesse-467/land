package routers

import (
	"land/controllers"
	"land/logger"
	"land/middlewares"
	"time"

	"github.com/gin-gonic/gin"
)

func SetRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(logger.GinLogger(),
		logger.GinRecovery(true),
		middlewares.RateLimit(2*time.Second, 50))

	auth := r.Group("/auth")
	{
		auth.POST("/login", controllers.LoginHandler)     // 登录
		auth.POST("/register", controllers.SignUpHandler) // 注册
		auth.POST("/logout", middlewares.JWTAuth(), controllers.LogoutHandler)
	}

	// 为后续路由启用JWT验证中间件
	v1 := r.Group("/api/v1")
	v1.Use(middlewares.JWTAuth())

	{
		// 社区相关
		v1.GET("/community", controllers.CommunityListController)       // 获取社区列表
		v1.GET("/community/:id", controllers.CommunityDetailController) // 获取社区详情

		// 帖子相关
		v1.GET("/post", controllers.GetPostListController)                           // 获取帖子列表
		v1.GET("/post/:id", controllers.PostDetailController)                        // 获取帖子详情
		v1.POST("/post", controllers.CreatePostController)                           // 创建帖子
		v1.PUT("/post", controllers.UpdatePostController)                            // 更新帖子（延迟双删）
		v1.PUT("/post/consistency", controllers.UpdatePostWithConsistencyController) // 更新帖子（强一致性）
		v1.POST("/vote", controllers.PostVoteController)                             // 帖子投票
		v1.GET("/posts2/", controllers.GetPostListHandler2)                          // 根据时间或分数获取帖子列表（优化版）

		// 评论相关
		v1.POST("/comment", controllers.CommentHandler)    // 评论
		v1.GET("/comment", controllers.CommentListHandler) // 评论列表

		// 管理相关
		v1.POST("/sync/viewcounts", controllers.SyncViewCountsHandler)  // 手动同步访问量
		v1.POST("/init/viewzset", controllers.InitPostViewZSetHandler)  // 初始化访问量有序集合
		v1.GET("/test/random-ttl", controllers.TestRandomTTLHandler)    // 测试随机TTL功能
		v1.DELETE("/post/:id/cache", controllers.ClearPostCacheHandler) // 清除指定帖子缓存
		v1.DELETE("/post/cache", controllers.ClearAllPostCacheHandler)  // 清除所有帖子缓存
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code": 404,
			"msg":  "Not Found",
		})
	})

	return r
}
