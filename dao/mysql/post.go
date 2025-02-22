package mysql

import (
	"land/models"
	"strings"

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
