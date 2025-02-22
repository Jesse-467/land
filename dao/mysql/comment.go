package mysql

import (
	"land/models"
	"time"

	"go.uber.org/zap"
)

func CreateComment(comment *models.Comment) error {
	comment.CreateTime = time.Now()
	comment.UpdateTime = time.Now()

	// 使用 GORM 插入数据
	if err := db.Create(comment).Error; err != nil {
		zap.L().Error("insert comment failed", zap.Error(err))
		return ErrorInsertFailed
	}
	return nil
}

func GetCommentListByIDs(ids []string) ([]*models.Comment, error) {
	commentList := make([]*models.Comment, 0)
	if err := db.Where("comment_id IN ?", ids).Find(&commentList).Error; err != nil {
		zap.L().Error("failed to get comment list", zap.Error(err))
		return nil, err
	}
	return commentList, nil
}
