package mysql

import (
	"land/models"

	"go.uber.org/zap"
)

func GetCommunityList() (list []*models.Community, err error) {
	list = make([]*models.Community, 0)
	err = db.Find(&list).Error
	if err != nil {
		zap.L().Warn("GetCommunityList failed", zap.Error(err))
	}
	return
}

func GetCommunityDetailByID(id uint64) (community *models.CommunityDetail, err error) {
	community = &models.CommunityDetail{}
	err = db.Where("community_id = ?", id).First(community).Error

	if err != nil {
		zap.L().Warn("GetCommunityDetailByID failed", zap.Error(err))
	}
	return community, err
}
