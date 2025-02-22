package redis

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// 推荐阅读
// 基于用户投票的相关算法：http://www.ruanyifeng.com/blog/algorithm/

// 本项目使用简化版的投票分数
// 投一票就加432分   86400/200  --> 200张赞成票可以给你的帖子续一天

/* 投票的几种情况：
   direction=1时，有两种情况：
   	1. 之前没有投过票，现在投赞成票    --> 更新分数和投票记录  差值的绝对值：1  +432
   	2. 之前投反对票，现在改投赞成票    --> 更新分数和投票记录  差值的绝对值：2  +432*2
   direction=0时，有两种情况：
   	1. 之前投过反对票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  +432
	2. 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  -432
   direction=-1时，有两种情况：
   	1. 之前没有投过票，现在投反对票    --> 更新分数和投票记录  差值的绝对值：1  -432
   	2. 之前投赞成票，现在改投反对票    --> 更新分数和投票记录  差值的绝对值：2  -432*2

   投票的限制：
   每个贴子自发表之日起一个星期之内允许用户投票，超过一个星期就不允许再投票了。
   	1. 到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
   	2. 到期之后删除那个 KeyPostVotedZSetPF
*/

const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 // 每一票值多少分
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeated   = errors.New("不允许重复投票")
)

func CreatePost(postID, communityID uint64) error {
	pipeline := client.TxPipeline()

	// 帖子时间
	pipeline.ZAdd(context.Background(), getRedisKey(KeyPostTimeZSet), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 帖子分数
	pipeline.ZAdd(context.Background(), getRedisKey(KeyPostScoreZSet), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 更新：把帖子id加到社区的set
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(context.Background(), cKey, postID)
	_, err := pipeline.Exec(context.Background())
	return err
}

func VoteForPost(userID, postID string, value float64) error {
	// 1. 先检查投票时间是否过期
	postTime := client.ZScore(context.Background(), getRedisKey(KeyPostTimeZSet), postID).Val()

	// 时间过期，返回错误
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}

	// 2. 检查是否重复投票
	voted := client.ZScore(context.Background(), getRedisKey(KeyPostVotedPF+postID), userID).Val()

	if voted == value {
		return ErrVoteRepeated
	}

	var op float64
	if value > voted {
		op = 1
	} else {
		op = -1
	}

	diff := math.Abs(value - voted)
	// 3. 更新帖子分数
	pipeline := client.TxPipeline()
	pipeline.ZIncrBy(context.Background(), getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID)

	// 4. 更新投票记录
	if value == 0 {
		pipeline.ZRem(context.Background(), getRedisKey(KeyPostVotedPF+postID), userID)
	} else {
		pipeline.ZAdd(context.Background(), getRedisKey(KeyPostVotedPF+postID), &redis.Z{
			Score:  value, // 赞成票还是反对票
			Member: userID,
		})
	}
	_, err := pipeline.Exec(context.Background())
	return err
}
