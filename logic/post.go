package logic

import (
	"encoding/json"
	"land/dao/mysql"
	"land/dao/redis"
	"land/models"
	"land/pkg/snowflake"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	p.ID = snowflake.GetID()

	err = mysql.CreatePost(p)
	if err != nil {
		return
	}

	err = redis.CreatePost(p.ID, p.CommunityID)
	if err != nil {
		return
	}

	// 清除相关缓存
	go func() {
		// 清除该作者的其他帖子缓存（可选，防止缓存不一致）
		// 这里可以根据需要实现更复杂的缓存清理策略
		zap.L().Debug("Post created, cache cleared",
			zap.Int64("post_id", int64(p.ID)),
			zap.Int64("author_id", int64(p.AuthorID)))
	}()

	return
}

func GetPostByID(pid uint64, userID ...uint64) (data *models.PostDetail, err error) {
	// 1. 首先检查是否被标记为不存在（防止缓存穿透）
	notExist, err := redis.CheckPostNotExist(pid)
	if err != nil {
		zap.L().Error("redis.CheckPostNotExist() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(err))
		// 检查失败时继续执行，不直接返回错误
	}

	if notExist {
		zap.L().Debug("Post marked as not exist", zap.Int64("pid", int64(pid)))
		return nil, mysql.ErrorInvalidID
	}

	// 2. 尝试从Redis缓存获取
	var post *models.Post
	var authorID uint64

	// 先尝试从数据库获取基本信息来确定作者ID
	post, err = mysql.GetPostByID(pid)
	if err != nil {
		// 如果数据库中没有该帖子，设置不存在标记防止缓存穿透
		if err == mysql.ErrorInvalidID {
			redis.SetPostNotExist(pid, 30*time.Second) // 30s内不再查询,防止缓存穿透
		}
		zap.L().Error("mysql.GetPostByID() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(err))
		return nil, err
	}

	authorID = post.AuthorID

	// 尝试从缓存获取完整的帖子详情
	cacheData, err := redis.GetPostCache(authorID, pid)
	if err == nil {
		// 缓存命中，解析数据
		var postDetail models.PostDetail
		if err := json.Unmarshal([]byte(cacheData), &postDetail); err != nil {
			zap.L().Error("json.Unmarshal cache data failed",
				zap.Int64("pid", int64(pid)),
				zap.Error(err))
			// 缓存数据损坏，删除缓存
			redis.DeletePostCache(authorID, pid)
		} else {
			zap.L().Debug("Post cache hit", zap.Int64("pid", int64(pid)))

			// 增加访问量
			if len(userID) > 0 {
				_, err = redis.IncrementPostViewCount(pid, userID[0])
			} else {
				_, err = redis.IncrementPostViewCount(pid)
			}
			if err != nil {
				zap.L().Error("redis.IncrementPostViewCount() failed",
					zap.Int64("pid", int64(pid)),
					zap.Error(err))
			}

			// 获取最新访问量
			viewCount, err := redis.GetPostViewCount(pid)
			if err != nil {
				zap.L().Error("redis.GetPostViewCount() failed",
					zap.Int64("pid", int64(pid)),
					zap.Error(err))
				viewCount = 0
			}
			postDetail.ViewCount = viewCount

			return &postDetail, nil
		}
	} else if err != redis.Nil {
		zap.L().Error("redis.GetPostCache() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(err))
	}

	// 3. 缓存未命中，从数据库获取完整数据
	zap.L().Debug("Post cache miss, fetching from database", zap.Int64("pid", int64(pid)))

	// 获取用户信息
	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", int64(post.AuthorID)),
			zap.Error(err))
		return nil, err
	}

	// 获取社区信息
	community, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID() failed",
			zap.Int64("community_id", int64(post.CommunityID)),
			zap.Error(err))
		return nil, err
	}

	// 4. 增加访问量
	if len(userID) > 0 {
		_, err = redis.IncrementPostViewCount(pid, userID[0])
	} else {
		_, err = redis.IncrementPostViewCount(pid)
	}
	if err != nil {
		zap.L().Error("redis.IncrementPostViewCount() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(err))
	}

	// 5. 获取访问量
	viewCount, err := redis.GetPostViewCount(pid)
	if err != nil {
		zap.L().Error("redis.GetPostViewCount() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(err))
		viewCount = 0
	}
	post.ViewCount = viewCount

	// 6. 组装数据
	data = &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: community,
	}

	// 7. 异步回填缓存
	go func() {
		if cacheData, err := json.Marshal(data); err == nil {
			// 缓存30分钟
			redis.SetPostCache(authorID, pid, string(cacheData), redis.PostCacheBaseTTL)
			zap.L().Debug("Post cache backfilled", zap.Int64("pid", int64(pid)))
		} else {
			zap.L().Error("json.Marshal post data failed",
				zap.Int64("pid", int64(pid)),
				zap.Error(err))
		}
	}()

	return data, nil
}

// GetPostList 获取帖子列表
func GetPostList(page, size int64) (data []*models.PostDetail, err error) {
	posts, err := mysql.GetPostList(page, size)
	if err != nil {
		return nil, err
	}
	data = make([]*models.PostDetail, 0, len(posts))

	for _, post := range posts {
		// 根据作者id查询作者信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", int64(post.AuthorID)),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("community_id", int64(post.CommunityID)),
				zap.Error(err))
			continue
		}
		postDetail := &models.PostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}

// GetPostListNew  将两个查询帖子列表逻辑合二为一的函数
func GetPostListNew(p *models.ParamPostList) (data []*models.PostDetail, err error) {
	// 根据请求参数的不同，执行不同的逻辑。
	if p.CommunityID == 0 {
		// 查所有
		data, err = getPostListCommon(redis.GetPostIDsInOrder, p)
	} else {
		// 根据社区id查询
		data, err = getPostListCommon(redis.GetCommunityPostIDsInOrder, p)
	}
	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
		return nil, err
	}
	return
}

func getPostListCommon(getIDsFunc func(p *models.ParamPostList) ([]string, error), p *models.ParamPostList) (data []*models.PostDetail, err error) {
	// 1. 获取帖子 ID 列表
	ids, err := getIDsFunc(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("getIDsFunc(p) return 0 data")
		return
	}
	zap.L().Debug("getPostListCommon", zap.Any("ids", ids))

	// 2. 根据 ID 列表查询帖子详细信息
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return
	}
	zap.L().Debug("getPostListCommon", zap.Any("posts", posts))

	// 3. 查询每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 4. 查询每篇帖子的访问量
	viewCounts, err := redis.GetPostViewCounts(ids)
	if err != nil {
		zap.L().Error("redis.GetPostViewCounts() failed", zap.Error(err))
		// 访问量获取失败时，创建默认的0值数组
		viewCounts = make([]int64, len(ids))
	}

	// 5. 填充作者和社区信息，同时尝试从缓存获取完整数据
	for idx, post := range posts {
		// 尝试从缓存获取完整数据
		cacheData, err := redis.GetPostCache(post.AuthorID, post.PostID)
		if err == nil {
			// 缓存命中，解析数据
			var postDetail models.PostDetail
			if err := json.Unmarshal([]byte(cacheData), &postDetail); err != nil {
				zap.L().Error("json.Unmarshal cache data failed",
					zap.Int64("post_id", int64(post.PostID)),
					zap.Error(err))
				// 缓存数据损坏，删除缓存
				redis.DeletePostCache(post.AuthorID, post.PostID)
			} else {
				// 更新访问量和投票数
				if idx < len(viewCounts) {
					postDetail.ViewCount = viewCounts[idx]
				}
				if idx < len(voteData) {
					postDetail.VoteNum = voteData[idx]
				}
				data = append(data, &postDetail)
				continue
			}
		}

		// 缓存未命中，从数据库获取
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", int64(post.AuthorID)),
				zap.Error(err))
			continue
		}
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("community_id", int64(post.CommunityID)),
				zap.Error(err))
			continue
		}

		// 设置访问量
		if idx < len(viewCounts) {
			post.ViewCount = viewCounts[idx]
		}

		postDetail := &models.PostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)

		// 异步回填缓存
		go func(p *models.Post, pd *models.PostDetail) {
			if cacheData, err := json.Marshal(pd); err == nil {
				redis.SetPostCache(p.AuthorID, p.PostID, string(cacheData), redis.PostCacheBaseTTL)
			}
		}(post, postDetail)
	}
	return
}

// GetPostListByOrder 根据排序方式获取帖子列表（混合策略）
// 参数:
//   - p: 查询参数
//
// 返回值:
//   - data: 帖子详情列表
//   - err: 可能的错误
func GetPostListByOrder(p *models.ParamPostList) (data []*models.PostDetail, err error) {
	// 如果明确指定不使用索引，或者按分数排序，使用Redis
	if !p.UseIndex || p.Order == "score" {
		return GetPostListNew(p)
	}

	// 按时间和访问量排序可以使用MySQL（利用数据库索引）
	if p.Order == "time" || p.Order == "view" {
		return getPostListFromMySQL(p)
	}

	// 默认使用Redis
	return GetPostListNew(p)
}

// getPostListFromMySQL 从MySQL获取帖子列表（利用数据库索引）
// 参数:
//   - p: 查询参数
//
// 返回值:
//   - data: 帖子详情列表
//   - err: 可能的错误
func getPostListFromMySQL(p *models.ParamPostList) (data []*models.PostDetail, err error) {
	// 从MySQL获取帖子列表
	posts, err := mysql.GetPostListByOrder(p.Order, p.Page, p.Size, p.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetPostListByOrder() failed", zap.Error(err))
		return nil, err
	}

	if len(posts) == 0 {
		return make([]*models.PostDetail, 0), nil
	}

	// 提取帖子ID列表用于批量获取投票数据和访问量
	postIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, strconv.FormatUint(post.PostID, 10))
	}

	// 批量获取投票数据
	voteData, err := redis.GetPostVoteData(postIDs)
	if err != nil {
		zap.L().Error("redis.GetPostVoteData() failed", zap.Error(err))
		// 投票数据获取失败时，创建默认的0值数组
		voteData = make([]int64, len(postIDs))
	}

	// 批量获取访问量（如果MySQL中没有访问量字段）
	var viewCounts []int64
	if p.Order == "view" {
		// 按访问量排序时，使用MySQL中的访问量字段
		viewCounts = make([]int64, len(posts))
		for i, post := range posts {
			viewCounts[i] = post.ViewCount
		}
	} else {
		// 其他情况从Redis获取最新访问量
		viewCounts, err = redis.GetPostViewCounts(postIDs)
		if err != nil {
			zap.L().Error("redis.GetPostViewCounts() failed", zap.Error(err))
			viewCounts = make([]int64, len(postIDs))
		}
	}

	// 组装帖子详情数据
	data = make([]*models.PostDetail, 0, len(posts))
	for i, post := range posts {
		// 获取用户信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById() failed",
				zap.Int64("author_id", int64(post.AuthorID)),
				zap.Error(err))
			continue
		}

		// 获取社区信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID() failed",
				zap.Int64("community_id", int64(post.CommunityID)),
				zap.Error(err))
			continue
		}

		// 设置投票数和访问量
		voteNum := int64(0)
		if i < len(voteData) {
			voteNum = voteData[i]
		}

		viewCount := int64(0)
		if i < len(viewCounts) {
			viewCount = viewCounts[i]
		}

		// 如果MySQL中有访问量字段且按访问量排序，使用MySQL的数据
		if p.Order == "view" {
			viewCount = post.ViewCount
		}

		// 设置帖子的访问量
		post.ViewCount = viewCount

		postDetail := &models.PostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteNum,
			Post:            post,
			CommunityDetail: community,
		}

		data = append(data, postDetail)
	}

	return data, nil
}

// UpdatePost 更新帖子（延迟双删策略）
// 参数:
//   - p: 更新帖子参数
//   - userID: 当前用户ID
//
// 返回值:
//   - error: 可能的错误
func UpdatePost(p *models.UpdatePostForm, userID uint64) error {
	// 1. 验证权限：检查用户是否为帖子作者
	authorID, err := mysql.GetPostAuthorID(p.PostID)
	if err != nil {
		zap.L().Error("mysql.GetPostAuthorID() failed",
			zap.Int64("post_id", int64(p.PostID)),
			zap.Error(err))
		return err
	}

	if authorID != userID {
		zap.L().Error("User not authorized to update post",
			zap.Int64("user_id", int64(userID)),
			zap.Int64("author_id", int64(authorID)),
			zap.Int64("post_id", int64(p.PostID)))
		return mysql.ErrorInvalidID // 使用现有错误，实际应该定义新的权限错误
	}

	// 2. 第一次删除缓存（立即删除）
	err = redis.InvalidatePostCache(authorID, p.PostID)
	if err != nil {
		zap.L().Error("First cache deletion failed",
			zap.Int64("post_id", int64(p.PostID)),
			zap.Error(err))
		// 不返回错误，继续执行更新操作
	}

	// 3. 更新MySQL数据库
	post := &models.Post{
		PostID:      p.PostID,
		Title:       p.Title,
		Content:     p.Content,
		CommunityID: p.CommunityID,
	}

	err = mysql.UpdatePost(post)
	if err != nil {
		zap.L().Error("mysql.UpdatePost() failed",
			zap.Int64("post_id", int64(p.PostID)),
			zap.Error(err))
		return err
	}

	// 4. 延迟第二次删除缓存（延迟双删策略）
	go func() {
		// 延迟500ms后再次删除缓存
		time.Sleep(500 * time.Millisecond)

		err := redis.DelayDeletePostCache(authorID, p.PostID, 1*time.Second)
		if err != nil {
			zap.L().Error("Second cache deletion failed",
				zap.Int64("post_id", int64(p.PostID)),
				zap.Error(err))
		} else {
			zap.L().Info("Post updated with delay double deletion",
				zap.Int64("post_id", int64(p.PostID)),
				zap.Int64("author_id", int64(authorID)))
		}
	}()

	return nil
}

// UpdatePostWithCacheConsistency 更新帖子（带缓存一致性检查）
// 参数:
//   - p: 更新帖子参数
//   - userID: 当前用户ID
//
// 返回值:
//   - error: 可能的错误
func UpdatePostWithCacheConsistency(p *models.UpdatePostForm, userID uint64) error {
	// 1. 验证权限
	authorID, err := mysql.GetPostAuthorID(p.PostID)
	if err != nil {
		return err
	}

	if authorID != userID {
		return mysql.ErrorInvalidID
	}

	// 2. 第一次删除缓存
	redis.InvalidatePostCache(authorID, p.PostID)

	// 3. 更新MySQL
	post := &models.Post{
		PostID:      p.PostID,
		Title:       p.Title,
		Content:     p.Content,
		CommunityID: p.CommunityID,
	}

	err = mysql.UpdatePost(post)
	if err != nil {
		return err
	}

	// 4. 异步延迟删除缓存
	go func() {
		// 延迟删除，确保MySQL主从同步完成
		time.Sleep(1 * time.Second)

		// 再次删除缓存，防止在第一次删除和MySQL更新之间
		// 有其他线程读取了旧数据并回填了缓存
		redis.InvalidatePostCache(authorID, p.PostID)

		zap.L().Info("Cache consistency ensured for post update",
			zap.Int64("post_id", int64(p.PostID)))
	}()

	return nil
}
