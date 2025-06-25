package mysql

import (
	"land/models"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CreatePost 创建新帖子
// 参数:
//   - p: 帖子信息
//
// 返回值:
//   - err: 可能的错误
func CreatePost(p *models.Post) error {
	if err := db.Create(p).Error; err != nil {
		zap.L().Error("CreatePost failed", zap.Error(err))
		return err
	}
	return nil
}

// GetPostById 根据帖子ID获取帖子信息
// 参数:
//   - pid: 帖子ID
//
// 返回值:
//   - post: 帖子信息
//   - err: 可能的错误
func GetPostByID(pid uint64) (post *models.Post, err error) {
	post = &models.Post{}
	if err = db.Where("post_id = ?", pid).First(post).Error; err != nil {
		// 检查是否是记录不存在的错误
		if err.Error() == "record not found" {
			return nil, ErrorInvalidID
		}
		return nil, err
	}
	return post, nil
}

// GetPostList 获取帖子列表
// 参数:
//   - page: 页码
//   - size: 每页数量
//
// 返回值:
//   - posts: 帖子列表
//   - err: 可能的错误
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	posts = make([]*models.Post, 0, 2)
	offset := (page - 1) * size

	err = db.Model(&models.Post{}).
		Select("post_id, title, content, author_id, community_id, create_time").
		Order("create_time DESC").
		Offset(int(offset)).
		Limit(int(size)).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostListByOrder 根据排序方式获取帖子列表（利用数据库索引）
// 参数:
//   - order: 排序方式 (time/view)
//   - page: 页码
//   - size: 每页数量
//   - communityID: 社区ID（可选，0表示所有社区）
//
// 返回值:
//   - posts: 帖子列表
//   - err: 可能的错误
func GetPostListByOrder(order string, page, size int64, communityID uint64) (posts []*models.Post, err error) {
	posts = make([]*models.Post, 0, size)
	offset := (page - 1) * size

	query := db.Model(&models.Post{}).
		Select("post_id, title, content, author_id, community_id, create_time, view_count")

	// 如果指定了社区ID，添加社区筛选条件
	if communityID > 0 {
		query = query.Where("community_id = ?", communityID)
	}

	// 根据排序方式设置排序字段，利用数据库索引
	switch order {
	case "time":
		query = query.Order("create_time DESC")
	case "view":
		query = query.Order("view_count DESC, create_time DESC") // 访问量相同时按时间排序
	default:
		query = query.Order("create_time DESC") // 默认按时间排序
	}

	err = query.Offset(int(offset)).Limit(int(size)).Find(&posts).Error
	if err != nil {
		zap.L().Error("GetPostListByOrder failed",
			zap.String("order", order),
			zap.Int64("page", page),
			zap.Int64("size", size),
			zap.Int64("community_id", int64(communityID)),
			zap.Error(err))
		return nil, err
	}

	return posts, nil
}

// GetPostCount 获取帖子总数
// 参数:
//   - communityID: 社区ID（可选，0表示所有社区）
//
// 返回值:
//   - count: 帖子总数
//   - err: 可能的错误
func GetPostCount(communityID uint64) (count int64, err error) {
	query := db.Model(&models.Post{})

	if communityID > 0 {
		query = query.Where("community_id = ?", communityID)
	}

	err = query.Count(&count).Error
	if err != nil {
		zap.L().Error("GetPostCount failed",
			zap.Int64("community_id", int64(communityID)),
			zap.Error(err))
		return 0, err
	}

	return count, nil
}

// GetPostListByIDs 根据帖子ID列表获取帖子信息
// 参数:
//   - ids: 帖子ID列表
//
// 返回值:
//   - postList: 帖子列表
//   - err: 可能的错误
func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
	idStr := strings.Join(ids, ",")
	postList = make([]*models.Post, 0, 2)

	err = db.Raw(`
        SELECT post_id, title, content, author_id, community_id, create_time
        FROM post
        WHERE post_id IN (?)
        ORDER BY FIND_IN_SET(post_id, ?)
    `, ids, idStr).Scan(&postList).Error

	if err != nil {
		return nil, err
	}

	return postList, nil
}

// UpdatePostViewCount 更新帖子访问量
// 参数:
//   - postID: 帖子ID
//   - viewCount: 访问量
//
// 返回值:
//   - err: 可能的错误
func UpdatePostViewCount(postID uint64, viewCount int64) error {
	err := db.Model(&models.Post{}).
		Where("post_id = ?", postID).
		Update("view_count", viewCount).Error
	if err != nil {
		zap.L().Error("UpdatePostViewCount failed",
			zap.Int64("post_id", int64(postID)),
			zap.Int64("view_count", viewCount),
			zap.Error(err))
		return err
	}
	return nil
}

// BatchUpdatePostViewCounts 批量更新帖子访问量
// 参数:
//   - viewCounts: 帖子ID和访问量的映射
//
// 返回值:
//   - err: 可能的错误
func BatchUpdatePostViewCounts(viewCounts map[uint64]int64) error {
	for postID, viewCount := range viewCounts {
		if err := UpdatePostViewCount(postID, viewCount); err != nil {
			return err
		}
	}
	return nil
}

// UpdatePost 更新帖子信息
// 参数:
//   - post: 帖子信息
//
// 返回值:
//   - err: 可能的错误
func UpdatePost(post *models.Post) error {
	// 更新帖子信息，只更新允许修改的字段
	err := db.Model(&models.Post{}).
		Where("post_id = ?", post.PostID).
		Updates(map[string]interface{}{
			"title":        post.Title,
			"content":      post.Content,
			"community_id": post.CommunityID,
			"update_time":  time.Now(),
		}).Error

	if err != nil {
		zap.L().Error("UpdatePost failed",
			zap.Int64("post_id", int64(post.PostID)),
			zap.Error(err))
		return err
	}

	return nil
}

// GetPostAuthorID 获取帖子作者ID
// 参数:
//   - postID: 帖子ID
//
// 返回值:
//   - authorID: 作者ID
//   - err: 可能的错误
func GetPostAuthorID(postID uint64) (authorID uint64, err error) {
	var post models.Post
	err = db.Select("author_id").Where("post_id = ?", postID).First(&post).Error
	if err != nil {
		return 0, err
	}
	return post.AuthorID, nil
}
