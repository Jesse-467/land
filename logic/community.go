package logic

import (
	"land/dao/mysql"
	"land/models"
)

func GetCommunityList() ([]*models.Community, error) {
	return mysql.GetCommunityList()
}

func GetCommunityDetail(id uint64) (*models.CommunityDetail, error) {
	return mysql.GetCommunityDetailByID(id)
}
