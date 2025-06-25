package controllers

import (
	"land/dao/mysql"
	"land/dao/redis"
	"land/logic"
	"land/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetPostListController(c *gin.Context) {
	page, size := GetPageInfo(c)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}
	ResSuccess(c, data)
	// 返回响应
}

func PostDetailController(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		zap.L().Error("get post detail failed with invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	// 获取当前用户ID（可选，用于防刷）
	var userID uint64
	if uid, err := GetCurrentUserID(c); err == nil {
		userID = uid
	}

	post, err := logic.GetPostByID(id, userID)
	if err != nil {
		zap.L().Error("logic.GetPostByID(id) failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, post)
}

func CreatePostController(c *gin.Context) {
	p := new(models.Post)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Debug("c.ShouldBindJSON(&p) failed", zap.Error(err))
		zap.L().Error("create post failed", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResError(c, CodeNeedLogin)
		return
	}
	p.AuthorID = userID

	if err = logic.CreatePost(p); err != nil {
		zap.L().Error("logic.CreatePost(p) failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, nil)
}

// GetPostListHandler2 升级版帖子列表接口
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间、分数或访问量排序查询帖子列表接口，支持分页（默认第1页，50条，最多100条），支持MySQL索引优化
// @Param page query int false "页码，默认为1"
// @Param size query int false "每页大小，默认为50，最大100"
// @Param order query string false "排序方式：time(时间倒序), score(分数倒序), view(访问量倒序)"
// @Param community_id query int false "社区ID，可选"
// @Param search query string false "搜索关键词，可选"
// @Param use_index query bool false "是否使用MySQL索引优化，默认true"
func GetPostListHandler2(c *gin.Context) {
	p := &models.ParamPostList{
		Page:     1,
		Size:     50,
		Order:    models.OrderTime,
		UseIndex: true, // 默认使用MySQL索引优化
	}

	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetPostListHandler2 with invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	// 设置默认值和限制
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size < 1 {
		p.Size = 50
	}
	if p.Size > 100 {
		p.Size = 100
	}

	data, err := logic.GetPostListByOrder(p) // 使用新的混合查询策略
	// 获取数据
	if err != nil {
		zap.L().Error("logic.GetPostListByOrder() failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResSuccess(c, data)

}

// SyncViewCountsHandler 手动同步访问量到MySQL
// @Summary 手动同步访问量
// @Description 手动触发Redis访问量数据同步到MySQL
func SyncViewCountsHandler(c *gin.Context) {
	// 这里可以添加权限检查，只有管理员才能调用，这里暂且不做权限检查
	// userID, err := GetCurrentUserID(c)
	// if err != nil {
	//     ResError(c, CodeNeedLogin)
	//     return
	// }

	syncService := logic.NewViewCountSyncService(0)
	err := syncService.ManualSync()
	if err != nil {
		zap.L().Error("Manual sync failed", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, gin.H{
		"message": "访问量同步完成",
	})
}

// ClearPostCacheHandler 清除帖子缓存
// @Summary 清除帖子缓存
// @Description 手动清除指定帖子的缓存
func ClearPostCacheHandler(c *gin.Context) {
	// 获取帖子ID参数
	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Invalid post ID", zap.String("post_id", postIDStr), zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	// 获取帖子信息以确定作者ID
	post, err := mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("Failed to get post", zap.Int64("post_id", int64(postID)), zap.Error(err))
		ResError(c, CodeNotFound)
		return
	}

	// 清除缓存
	err = redis.DeletePostCache(post.AuthorID, postID)
	if err != nil {
		zap.L().Error("Failed to delete post cache",
			zap.Int64("post_id", int64(postID)),
			zap.Int64("author_id", int64(post.AuthorID)),
			zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, gin.H{
		"message": "帖子缓存已清除",
		"post_id": postID,
	})
}

// ClearAllPostCacheHandler 清除所有帖子缓存
// @Summary 清除所有帖子缓存
// @Description 手动清除所有帖子的缓存
func ClearAllPostCacheHandler(c *gin.Context) {
	// 这里可以实现清除所有缓存的逻辑
	// 由于Redis没有直接清除所有缓存的命令，可以通过模式匹配来实现

	ResSuccess(c, gin.H{
		"message": "所有帖子缓存清除功能待实现",
	})
}

// UpdatePostController 更新帖子
// @Summary 更新帖子
// @Description 更新帖子信息，采用延迟双删策略保证缓存一致性
func UpdatePostController(c *gin.Context) {
	// 获取当前用户ID
	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResError(c, CodeNeedLogin)
		return
	}

	// 绑定请求参数
	p := new(models.UpdatePostForm)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdatePostController: invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	// 调用业务逻辑
	err = logic.UpdatePost(p, userID)
	if err != nil {
		zap.L().Error("logic.UpdatePost() failed", zap.Error(err))
		if err == mysql.ErrorInvalidID {
			ResError(c, CodeUnauthorized)
		} else {
			ResError(c, CodeServerBusy)
		}
		return
	}

	ResSuccess(c, gin.H{
		"message": "帖子更新成功",
		"post_id": p.PostID,
	})
}

// UpdatePostWithConsistencyController 更新帖子（强一致性版本）
// @Summary 更新帖子（强一致性）
// @Description 更新帖子信息，采用强一致性策略保证缓存一致性
func UpdatePostWithConsistencyController(c *gin.Context) {
	// 获取当前用户ID
	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResError(c, CodeNeedLogin)
		return
	}

	// 绑定请求参数
	p := new(models.UpdatePostForm)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("UpdatePostWithConsistencyController: invalid params", zap.Error(err))
		ResError(c, CodeInvalidParams)
		return
	}

	// 调用业务逻辑
	err = logic.UpdatePostWithCacheConsistency(p, userID)
	if err != nil {
		zap.L().Error("logic.UpdatePostWithCacheConsistency() failed", zap.Error(err))
		if err == mysql.ErrorInvalidID {
			ResError(c, CodeUnauthorized)
		} else {
			ResError(c, CodeServerBusy)
		}
		return
	}

	ResSuccess(c, gin.H{
		"message": "帖子更新完成（强一致性）",
		"post_id": p.PostID,
	})
}

// InitPostViewZSetHandler 初始化帖子访问量有序集合
// @Summary 初始化访问量排序
// @Description 手动初始化Redis中的帖子访问量有序集合，用于按访问量排序功能
func InitPostViewZSetHandler(c *gin.Context) {
	// 这里可以添加权限检查，只有管理员才能调用
	// userID, err := GetCurrentUserID(c)
	// if err != nil {
	//     ResError(c, CodeNeedLogin)
	//     return
	// }

	// 获取所有访问量数据
	viewCounts, err := redis.GetAllPostViewCounts()
	if err != nil {
		zap.L().Error("Failed to get all post view counts", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	// 初始化访问量有序集合
	err = redis.InitPostViewZSet(viewCounts)
	if err != nil {
		zap.L().Error("Failed to init post view ZSet", zap.Error(err))
		ResError(c, CodeServerBusy)
		return
	}

	ResSuccess(c, gin.H{
		"message":     "访问量有序集合初始化完成",
		"posts_count": len(viewCounts),
	})
}

// TestRandomTTLHandler 测试随机TTL生成功能
// @Summary 测试随机TTL
// @Description 测试随机TTL生成功能，验证缓存雪崩防护
func TestRandomTTLHandler(c *gin.Context) {
	// 获取测试参数
	baseTTLStr := c.Query("base_ttl")
	jitterPercentStr := c.Query("jitter_percent")
	iterationsStr := c.Query("iterations")

	// 解析参数
	baseTTL, err := time.ParseDuration(baseTTLStr)
	if err != nil || baseTTL <= 0 {
		baseTTL = 30 * time.Minute // 默认30分钟
	}

	jitterPercent, err := strconv.Atoi(jitterPercentStr)
	if err != nil || jitterPercent < 0 || jitterPercent > 100 {
		jitterPercent = 20 // 默认20%
	}

	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil || iterations <= 0 || iterations > 1000 {
		iterations = 100 // 默认100次
	}

	// 执行测试
	results := redis.TestRandomTTL(baseTTL, jitterPercent, iterations)

	// 计算统计信息
	minTTL := results[0]
	maxTTL := results[0]
	for _, ttl := range results {
		if ttl < minTTL {
			minTTL = ttl
		}
		if ttl > maxTTL {
			maxTTL = ttl
		}
	}

	expectedMin := time.Duration(float64(baseTTL) * (1 - float64(jitterPercent)/100))
	expectedMax := time.Duration(float64(baseTTL) * (1 + float64(jitterPercent)/100))

	ResSuccess(c, gin.H{
		"test_params": gin.H{
			"base_ttl":       baseTTL.String(),
			"jitter_percent": jitterPercent,
			"iterations":     iterations,
		},
		"results": gin.H{
			"min_ttl":      minTTL.String(),
			"max_ttl":      maxTTL.String(),
			"expected_min": expectedMin.String(),
			"expected_max": expectedMax.String(),
			"sample_ttls":  results[:10], // 返回前10个样本
		},
		"message": "随机TTL测试完成",
	})
}
