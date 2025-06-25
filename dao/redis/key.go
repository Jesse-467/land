package redis

import (
	"time"
)

// Redis键定义
// 注意：使用命名空间方式定义Redis键，以便于查询和拆分

// 常量定义
const (
	// Prefix 项目前缀
	Prefix = "land:"

	// KeyPostTimeZSet 帖子时间有序集合
	// 类型：zset
	// 用途：存储帖子ID及其发布时间
	KeyPostTimeZSet = "post:time"

	// KeyPostScoreZSet 帖子分数有序集合
	// 类型：zset
	// 用途：存储帖子ID及其投票分数
	KeyPostScoreZSet = "post:score"

	// KeyPostVotedPF 用户投票记录
	// 类型：zset
	// 用途：记录用户对帖子的投票类型
	KeyPostVotedPF = "post:voted:"

	// KeyPostIDSet 帖子ID集合
	// 类型：set
	// 用途：保存每个分区下帖子id
	KeyCommunitySetPF = "community:"

	// KeyPostViewCountPF 帖子访问量
	// 类型：string
	// 用途：存储帖子的访问量计数
	KeyPostViewCountPF = "post:view:"

	// KeyPostViewSetPF 帖子访问记录
	// 类型：set
	// 用途：记录用户是否已访问过该帖子（防刷）
	KeyPostViewSetPF = "post:viewed:"

	// KeyPostCachePF 帖子缓存
	// 类型：string
	// 用途：存储帖子的完整信息缓存
	KeyPostCachePF = "post:cache:"

	// KeyPostNotExistPF 帖子不存在标记
	// 类型：string
	// 用途：防止缓存穿透，标记不存在的帖子
	KeyPostNotExistPF = "post:notexist:"

	// KeyPostViewZSet 帖子访问量有序集合
	// 类型：zset
	// 用途：存储帖子ID及其访问量
	KeyPostViewZSet = "post:view"
)

// getRedisKey 获取完整的Redis键
// 参数：
//   - key: 键名
//
// 返回值：
//   - 带有项目前缀的完整键名
func getRedisKey(key string) string {
	return Prefix + key
}

// 缓存TTL配置常量
const (
	// 帖子缓存TTL配置
	PostCacheBaseTTL       = 30 * time.Minute // 帖子缓存基础TTL
	PostCacheJitterPercent = 20               // 帖子缓存随机抖动百分比

	// 帖子不存在标记TTL配置
	PostNotExistBaseTTL       = 5 * time.Minute // 帖子不存在标记基础TTL
	PostNotExistJitterPercent = 15              // 帖子不存在标记随机抖动百分比

	// 访问量缓存TTL配置
	ViewCountBaseTTL       = 7 * 24 * time.Hour // 访问量缓存基础TTL
	ViewCountJitterPercent = 15                 // 访问量缓存随机抖动百分比

	// 用户访问记录TTL配置
	UserViewedBaseTTL       = 24 * time.Hour // 用户访问记录基础TTL
	UserViewedJitterPercent = 10             // 用户访问记录随机抖动百分比

	// 社区缓存TTL配置
	CommunityCacheBaseTTL       = 60 * time.Second // 社区缓存基础TTL
	CommunityCacheJitterPercent = 25               // 社区缓存随机抖动百分比
)
