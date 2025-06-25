package models

import "time"

type Post struct {
	ID          uint64    `json:"id"`
	PostID      uint64    `json:"post_id"`
	AuthorID    uint64    `json:"author_id"`
	CommunityID uint64    `json:"community_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Status      uint8     `json:"status"`
	ViewCount   int64     `json:"view_count"` // 访问量
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func (p *Post) TableName() string {
	return "post"
}

// 帖子返回的详细信息
type PostDetail struct {
	AuthorName       string             `json:"author_name"`
	VoteNum          int64              `json:"vote_num"`
	*Post                               // 嵌入帖子基本信息
	*CommunityDetail `json:"community"` // 嵌入社区信息
}

type Page struct {
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Size  int64 `json:"size"`
}

type PostDetailRes struct {
	Page Page          `json:"page"`
	List []*PostDetail `json:"list"`
}
