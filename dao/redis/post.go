package redis

import (
	"context"
	"land/dao/mysql"
	"land/models"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// generateRandomTTL 生成随机TTL，防止缓存雪崩
// 参数:
//   - baseTTL: 基础TTL时间
//   - jitterPercent: 随机抖动百分比（0-100）
//
// 返回值:
//   - time.Duration: 随机TTL时间
func generateRandomTTL(baseTTL time.Duration, jitterPercent int) time.Duration {
	if jitterPercent <= 0 || jitterPercent > 100 {
		return baseTTL
	}

	// 计算随机抖动范围
	jitterRange := float64(baseTTL) * float64(jitterPercent) / 100.0

	// 生成随机数种子（基于当前时间）
	rand.Seed(time.Now().UnixNano())

	// 生成随机抖动时间（正负范围）
	randomJitter := rand.Float64()*jitterRange*2 - jitterRange

	// 计算最终TTL
	finalTTL := float64(baseTTL) + randomJitter

	// 确保TTL不为负数
	if finalTTL < 0 {
		finalTTL = float64(baseTTL) * 0.1 // 最小为原TTL的10%
	}

	return time.Duration(finalTTL)
}

func getIDsFormKey(key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1
	// 3. ZREVRANGE 按分数从大到小的顺序查询指定数量的元素
	return client.ZRevRange(context.Background(), key, start, end).Result()
}

func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从redis获取id
	// 1. 根据用户请求中携带的order参数确定要查询的redis key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	} else if p.Order == models.OrderView {
		key = getRedisKey(KeyPostViewZSet)
	}
	// 2. 确定查询的索引起始点
	return getIDsFormKey(key, p.Page, p.Size)
}

// GetPostVoteData 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	//data = make([]int64, 0, len(ids))

	// 使用pipeline一次发送多条命令,减少RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedPF + id)
		pipeline.ZCount(context.Background(), key, "1", "1")
	}

	cmders, err := pipeline.Exec(context.Background())
	if err != nil {
		return nil, err
	}

	data = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

// GetCommunityPostIDsInOrder 根据社区id查询社区帖子的id列表
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	} else if p.Order == models.OrderView {
		orderKey = getRedisKey(KeyPostViewZSet)
	}

	// 使用 zinterstore 把分区的帖子set与帖子分数的 zset 生成一个新的zset
	// 针对新的zset 按之前的逻辑取数据

	// 社区的key
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(p.CommunityID)))

	// 利用缓存key减少zinterstore的开销，提高效率
	key := orderKey + strconv.Itoa(int(p.CommunityID))
	zap.L().Debug("key", zap.Any("orderkey+comkeys", key))

	if client.Exists(context.Background(), key).Val() < 1 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(context.Background(), key, &redis.ZStore{
			Keys:      []string{cKey, orderKey},
			Aggregate: "MAX",
		}) // zinterstore 计算

		// 生成随机TTL，防止缓存雪崩
		communityRandomTTL := generateRandomTTL(CommunityCacheBaseTTL, CommunityCacheJitterPercent)
		pipeline.Expire(context.Background(), key, communityRandomTTL)

		_, err := pipeline.Exec(context.Background())
		if err != nil {
			return nil, err
		}
	}
	// 存在的话就直接根据key查询ids
	return getIDsFormKey(key, p.Page, p.Size)
}

// IncrementPostViewCount 增加帖子访问量
// 参数:
//   - postID: 帖子ID
//   - userID: 用户ID（可选，用于防刷）
//
// 返回值:
//   - newCount: 新的访问量
//   - err: 可能的错误
func IncrementPostViewCount(postID uint64, userID ...uint64) (newCount int64, err error) {
	ctx := context.Background()
	viewCountKey := getRedisKey(KeyPostViewCountPF + strconv.FormatUint(postID, 10))
	viewZSetKey := getRedisKey(KeyPostViewZSet)

	// 如果有用户ID，检查是否已访问过（防刷）
	if len(userID) > 0 && userID[0] > 0 {
		viewedKey := getRedisKey(KeyPostViewSetPF + strconv.FormatUint(postID, 10))
		userIDStr := strconv.FormatUint(userID[0], 10)

		// 检查用户是否已访问过该帖子
		exists, err := client.SIsMember(ctx, viewedKey, userIDStr).Result()
		if err != nil {
			return 0, err
		}

		// 如果已访问过，不增加访问量
		if exists {
			// 返回当前访问量
			count, err := client.Get(ctx, viewCountKey).Int64()
			if err == redis.Nil {
				return 0, nil
			}
			return count, err
		}

		// 记录用户已访问并增加访问量
		pipeline := client.Pipeline()
		pipeline.SAdd(ctx, viewedKey, userIDStr)

		// 生成随机TTL，防止缓存雪崩
		viewedRandomTTL := generateRandomTTL(UserViewedBaseTTL, UserViewedJitterPercent)
		pipeline.Expire(ctx, viewedKey, viewedRandomTTL)

		pipeline.Incr(ctx, viewCountKey)

		// 生成随机TTL，防止缓存雪崩
		viewCountRandomTTL := generateRandomTTL(ViewCountBaseTTL, ViewCountJitterPercent)
		pipeline.Expire(ctx, viewCountKey, viewCountRandomTTL)

		cmders, err := pipeline.Exec(ctx)
		if err != nil {
			return 0, err
		}

		// 获取新的访问量
		newCount = cmders[2].(*redis.IntCmd).Val()

		// 更新访问量有序集合
		go func() {
			client.ZAdd(ctx, viewZSetKey, &redis.Z{
				Score:  float64(newCount),
				Member: strconv.FormatUint(postID, 10),
			})
		}()

		return newCount, nil
	}

	// 没有用户ID，直接增加访问量
	newCount, err = client.Incr(ctx, viewCountKey).Result()
	if err != nil {
		return 0, err
	}

	// 设置过期时间（添加随机TTL）
	viewCountRandomTTL := generateRandomTTL(ViewCountBaseTTL, ViewCountJitterPercent)
	client.Expire(ctx, viewCountKey, viewCountRandomTTL)

	// 更新访问量有序集合
	go func() {
		client.ZAdd(ctx, viewZSetKey, &redis.Z{
			Score:  float64(newCount),
			Member: strconv.FormatUint(postID, 10),
		})
	}()

	return newCount, nil
}

// GetPostViewCount 获取帖子访问量
// 参数:
//   - postID: 帖子ID
//
// 返回值:
//   - count: 访问量
//   - err: 可能的错误
func GetPostViewCount(postID uint64) (count int64, err error) {
	ctx := context.Background()
	viewCountKey := getRedisKey(KeyPostViewCountPF + strconv.FormatUint(postID, 10))

	count, err = client.Get(ctx, viewCountKey).Int64()
	if err == redis.Nil {
		return 0, nil // 没有访问记录，返回0
	}
	return count, err
}

// GetPostViewCounts 批量获取帖子访问量
// 参数:
//   - postIDs: 帖子ID列表
//
// 返回值:
//   - counts: 访问量列表
//   - err: 可能的错误
func GetPostViewCounts(postIDs []string) (counts []int64, err error) {
	ctx := context.Background()

	// 使用pipeline批量获取
	pipeline := client.Pipeline()
	for _, id := range postIDs {
		viewCountKey := getRedisKey(KeyPostViewCountPF + id)
		pipeline.Get(ctx, viewCountKey)
	}

	cmders, err := pipeline.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	counts = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		if cmder.Err() == redis.Nil {
			counts = append(counts, 0)
		} else {
			count, err := cmder.(*redis.StringCmd).Int64()
			if err != nil {
				counts = append(counts, 0)
			} else {
				counts = append(counts, count)
			}
		}
	}

	return counts, nil
}

// GetAllPostViewCounts 获取所有帖子的访问量
// 返回值:
//   - viewCounts: 帖子ID和访问量的映射
//   - err: 可能的错误
func GetAllPostViewCounts() (viewCounts map[uint64]int64, err error) {
	ctx := context.Background()
	viewCounts = make(map[uint64]int64)

	// 获取所有访问量键
	pattern := getRedisKey(KeyPostViewCountPF) + "*"
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	// 批量获取所有访问量
	pipeline := client.Pipeline()
	for _, key := range keys {
		pipeline.Get(ctx, key)
	}

	cmders, err := pipeline.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// 解析结果
	for i, cmder := range cmders {
		if cmder.Err() == redis.Nil {
			continue
		}

		count, err := cmder.(*redis.StringCmd).Int64()
		if err != nil {
			continue
		}

		// 从key中提取帖子ID
		key := keys[i]
		postIDStr := strings.TrimPrefix(key, getRedisKey(KeyPostViewCountPF))
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			continue
		}

		viewCounts[postID] = count
	}

	return viewCounts, nil
}

// SyncViewCountsToMySQL 同步访问量到MySQL
// 参数:
//   - viewCounts: 帖子ID和访问量的映射
//
// 返回值:
//   - err: 可能的错误
func SyncViewCountsToMySQL(viewCounts map[uint64]int64) error {
	if len(viewCounts) == 0 {
		return nil
	}

	// 调用MySQL批量更新
	return mysql.BatchUpdatePostViewCounts(viewCounts)
}

// CleanExpiredViewRecords 清理过期的访问记录
// 参数:
//   - postID: 帖子ID
//
// 返回值:
//   - err: 可能的错误
func CleanExpiredViewRecords(postID uint64) error {
	ctx := context.Background()
	viewedKey := getRedisKey(KeyPostViewSetPF + strconv.FormatUint(postID, 10))

	// 检查访问记录是否过期，如果过期则删除
	ttl, err := client.TTL(ctx, viewedKey).Result()
	if err != nil {
		return err
	}

	// 如果TTL为-1（永不过期）或-2（键不存在），则删除
	if ttl == -1 || ttl == -2 {
		client.Del(ctx, viewedKey)
	}

	return nil
}

// GetPostCacheKey 生成帖子缓存键
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//
// 返回值:
//   - string: 缓存键
func GetPostCacheKey(authorID, postID uint64) string {
	return getRedisKey(KeyPostCachePF) + strconv.FormatUint(authorID, 10) + ":" + strconv.FormatUint(postID, 10)
}

// GetPostNotExistKey 生成帖子不存在标记键
// 参数:
//   - postID: 帖子ID
//
// 返回值:
//   - string: 不存在标记键
func GetPostNotExistKey(postID uint64) string {
	return getRedisKey(KeyPostNotExistPF) + strconv.FormatUint(postID, 10)
}

// SetPostCache 设置帖子缓存
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//   - postData: 帖子数据（JSON字符串）
//   - expireTime: 过期时间
//
// 返回值:
//   - error: 可能的错误
func SetPostCache(authorID, postID uint64, postData string, expireTime time.Duration) error {
	ctx := context.Background()
	cacheKey := GetPostCacheKey(authorID, postID)

	// 生成随机TTL，防止缓存雪崩（使用传入的expireTime作为基础TTL）
	randomTTL := generateRandomTTL(expireTime, PostCacheJitterPercent)

	err := client.Set(ctx, cacheKey, postData, randomTTL).Err()
	if err != nil {
		zap.L().Error("SetPostCache failed",
			zap.Int64("author_id", int64(authorID)),
			zap.Int64("post_id", int64(postID)),
			zap.Duration("original_ttl", expireTime),
			zap.Duration("random_ttl", randomTTL),
			zap.Error(err))
		return err
	}

	zap.L().Debug("Post cache set with random TTL",
		zap.Int64("author_id", int64(authorID)),
		zap.Int64("post_id", int64(postID)),
		zap.Duration("original_ttl", expireTime),
		zap.Duration("random_ttl", randomTTL))

	return nil
}

// GetPostCache 获取帖子缓存
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//
// 返回值:
//   - string: 帖子数据（JSON字符串）
//   - error: 可能的错误
func GetPostCache(authorID, postID uint64) (string, error) {
	ctx := context.Background()
	cacheKey := GetPostCacheKey(authorID, postID)

	data, err := client.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return "", redis.Nil // 缓存不存在
	}
	if err != nil {
		zap.L().Error("GetPostCache failed",
			zap.Int64("author_id", int64(authorID)),
			zap.Int64("post_id", int64(postID)),
			zap.Error(err))
		return "", err
	}

	return data, nil
}

// SetPostNotExist 设置帖子不存在标记（防止缓存穿透）
// 参数:
//   - postID: 帖子ID
//   - expireTime: 过期时间
//
// 返回值:
//   - error: 可能的错误
func SetPostNotExist(postID uint64, expireTime time.Duration) error {
	ctx := context.Background()
	notExistKey := GetPostNotExistKey(postID)

	// 生成随机TTL，防止缓存雪崩（使用传入的expireTime作为基础TTL）
	randomTTL := generateRandomTTL(expireTime, PostNotExistJitterPercent)

	err := client.Set(ctx, notExistKey, "1", randomTTL).Err()
	if err != nil {
		zap.L().Error("SetPostNotExist failed",
			zap.Int64("post_id", int64(postID)),
			zap.Duration("original_ttl", expireTime),
			zap.Duration("random_ttl", randomTTL),
			zap.Error(err))
		return err
	}

	zap.L().Debug("Post not exist cache set with random TTL",
		zap.Int64("post_id", int64(postID)),
		zap.Duration("original_ttl", expireTime),
		zap.Duration("random_ttl", randomTTL))

	return nil
}

// CheckPostNotExist 检查帖子是否被标记为不存在
// 参数:
//   - postID: 帖子ID
//
// 返回值:
//   - bool: 是否不存在
//   - error: 可能的错误
func CheckPostNotExist(postID uint64) (bool, error) {
	ctx := context.Background()
	notExistKey := GetPostNotExistKey(postID)

	exists, err := client.Exists(ctx, notExistKey).Result()
	if err != nil {
		zap.L().Error("CheckPostNotExist failed",
			zap.Int64("post_id", int64(postID)),
			zap.Error(err))
		return false, err
	}

	return exists > 0, nil
}

// DeletePostCache 删除帖子缓存
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//
// 返回值:
//   - error: 可能的错误
func DeletePostCache(authorID, postID uint64) error {
	ctx := context.Background()
	cacheKey := GetPostCacheKey(authorID, postID)

	err := client.Del(ctx, cacheKey).Err()
	if err != nil {
		zap.L().Error("DeletePostCache failed",
			zap.Int64("author_id", int64(authorID)),
			zap.Int64("post_id", int64(postID)),
			zap.Error(err))
		return err
	}

	return nil
}

// BatchDeletePostCache 批量删除帖子缓存
// 参数:
//   - posts: 帖子列表
//
// 返回值:
//   - error: 可能的错误
func BatchDeletePostCache(posts []*models.Post) error {
	if len(posts) == 0 {
		return nil
	}

	ctx := context.Background()
	pipeline := client.Pipeline()

	for _, post := range posts {
		cacheKey := GetPostCacheKey(post.AuthorID, post.PostID)
		pipeline.Del(ctx, cacheKey)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("BatchDeletePostCache failed", zap.Error(err))
		return err
	}

	return nil
}

// DelayDeletePostCache 延迟删除帖子缓存（延迟双删策略）
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//   - delayTime: 延迟时间
//
// 返回值:
//   - error: 可能的错误
func DelayDeletePostCache(authorID, postID uint64, delayTime time.Duration) error {
	ctx := context.Background()
	cacheKey := GetPostCacheKey(authorID, postID)

	// 使用Redis的EXPIRE命令设置延迟删除
	err := client.Expire(ctx, cacheKey, delayTime).Err()
	if err != nil {
		zap.L().Error("DelayDeletePostCache failed",
			zap.Int64("author_id", int64(authorID)),
			zap.Int64("post_id", int64(postID)),
			zap.Duration("delay_time", delayTime),
			zap.Error(err))
		return err
	}

	zap.L().Debug("Post cache scheduled for deletion",
		zap.Int64("author_id", int64(authorID)),
		zap.Int64("post_id", int64(postID)),
		zap.Duration("delay_time", delayTime))

	return nil
}

// BatchDelayDeletePostCache 批量延迟删除帖子缓存
// 参数:
//   - posts: 帖子列表
//   - delayTime: 延迟时间
//
// 返回值:
//   - error: 可能的错误
func BatchDelayDeletePostCache(posts []*models.Post, delayTime time.Duration) error {
	if len(posts) == 0 {
		return nil
	}

	ctx := context.Background()
	pipeline := client.Pipeline()

	for _, post := range posts {
		cacheKey := GetPostCacheKey(post.AuthorID, post.PostID)
		pipeline.Expire(ctx, cacheKey, delayTime)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("BatchDelayDeletePostCache failed", zap.Error(err))
		return err
	}

	return nil
}

// InvalidatePostCache 使帖子缓存失效
// 参数:
//   - authorID: 作者ID
//   - postID: 帖子ID
//
// 返回值:
//   - err: 可能的错误
func InvalidatePostCache(authorID, postID uint64) error {
	ctx := context.Background()
	cacheKey := GetPostCacheKey(authorID, postID)

	err := client.Del(ctx, cacheKey).Err()
	if err != nil {
		zap.L().Error("InvalidatePostCache failed",
			zap.Int64("post_id", int64(postID)),
			zap.Int64("author_id", int64(authorID)),
			zap.Error(err))
		return err
	}

	return nil
}

// InitPostViewZSet 初始化帖子访问量有序集合
// 参数:
//   - viewCounts: 帖子ID和访问量的映射
//
// 返回值:
//   - err: 可能的错误
func InitPostViewZSet(viewCounts map[uint64]int64) error {
	ctx := context.Background()
	viewZSetKey := getRedisKey(KeyPostViewZSet)

	// 批量添加访问量数据到有序集合
	pipeline := client.Pipeline()
	for postID, viewCount := range viewCounts {
		pipeline.ZAdd(ctx, viewZSetKey, &redis.Z{
			Score:  float64(viewCount),
			Member: strconv.FormatUint(postID, 10),
		})
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("InitPostViewZSet failed", zap.Error(err))
		return err
	}

	zap.L().Info("Post view ZSet initialized", zap.Int("count", len(viewCounts)))
	return nil
}

// TestRandomTTL 测试随机TTL生成功能
// 参数:
//   - baseTTL: 基础TTL时间
//   - jitterPercent: 随机抖动百分比
//   - iterations: 测试迭代次数
//
// 返回值:
//   - []time.Duration: 生成的随机TTL列表
func TestRandomTTL(baseTTL time.Duration, jitterPercent int, iterations int) []time.Duration {
	results := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		randomTTL := generateRandomTTL(baseTTL, jitterPercent)
		results = append(results, randomTTL)
	}

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

	zap.L().Info("Random TTL test completed",
		zap.Duration("base_ttl", baseTTL),
		zap.Int("jitter_percent", jitterPercent),
		zap.Int("iterations", iterations),
		zap.Duration("min_ttl", minTTL),
		zap.Duration("max_ttl", maxTTL),
		zap.Duration("expected_min", time.Duration(float64(baseTTL)*(1-float64(jitterPercent)/100))),
		zap.Duration("expected_max", time.Duration(float64(baseTTL)*(1+float64(jitterPercent)/100))))

	return results
}
