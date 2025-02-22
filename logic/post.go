package logic

import (
	"land/dao/mysql"
	"land/dao/redis"
	"land/models"
	"land/pkg/snowflake"

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

	return
}

func GetPostByID(pid uint64) (data *models.PostDetail, err error) {
	post, e := mysql.GetPostByID(pid)
	if e != nil {
		zap.L().Error("mysql.GetPostByID() failed",
			zap.Int64("pid", int64(pid)),
			zap.Error(e))
		return
	}

	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", int64(post.AuthorID)),
			zap.Error(err))
		return
	}
	// 根据社区id查询社区详细信息
	community, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("community_id", int64(post.CommunityID)),
			zap.Error(err))
		return
	}
	// 接口数据拼接
	data = &models.PostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: community,
	}
	return
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

	// 4. 填充作者和社区信息
	for idx, post := range posts {
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
		postDetail := &models.PostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}
