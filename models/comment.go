package models

import "time"

type Comment struct {
	ID         uint64    `json:"id"`
	CommentID  uint64    `json:"comment_id"`
	PostID     uint64    `json:"post_id"`
	AuthorID   uint64    `json:"author_id"`
	ParentID   uint64    `json:"parent_id"`
	Content    string    `json:"content"`
	Status     uint8     `json:"status"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (c *Comment) TableName() string {
	return "comment"
}
