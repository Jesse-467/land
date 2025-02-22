package models

import "time"

type Community struct {
	ID            uint64    `json:"id"`
	CommunityID   uint64    `json:"community_id"`
	CommunityName string    `json:"community_name"`
	Introduction  string    `json:"introduction"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
}

type CommunityDetail struct {
	CommunityID   uint64    `json:"community_id" `
	CommunityName string    `json:"community_name" `
	Introduction  string    `json:"introduction,omitempty" ` // omitempty 当Introduction为空时不展示
	CreateTime    time.Time `json:"create_time"`
}

func (c *Community) TableName() string {
	return "community"
}

func (c *CommunityDetail) TableName() string {
	return "community"
}
