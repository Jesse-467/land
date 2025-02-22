package logic

import (
	"land/dao/redis"
	"land/models"
	"strconv"

	"go.uber.org/zap"
)

// VoteForPost 处理用户对帖子的投票
// 参数：
//   - id: 用户ID
//   - p: 投票参数
//
// 返回值：
//   - error: 可能发生的错误
func VoteForPost(id uint64, p *models.ParamVoteData) error {
	// 记录调试日志
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", int64(id)),
		zap.String("postID", p.PostID),
		zap.Int8("direction", p.Direction))

	// 调用Redis处理投票
	return redis.VoteForPost(strconv.Itoa(int(id)), p.PostID, float64(p.Direction))
}
