package models

const (
	OrderTime  = "time"
	OrderScore = "score"
	OrderView  = "view" // 按访问量排序
)

// type RegisterForm struct {
// 	Username        string `form:"username" binding:"required"`
// 	Password        string `form:"password" binding:"required"`
// 	Email           string `form:"email" binding:"required"`
// 	ConfirmPassword string `form:"confirm_password" binding:"required,eqfield=Password"`
// 	// 0:unknown, 1:male, 2:female
// 	Gender string `form:"gender" binding:"oneof=0 1 2"`
// }

type SignUpForm struct {
	UserName   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
	Email      string `json:"email"`
}

type LoginForm struct {
	UserName string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type ParamVoteData struct {
	PostID    string `form:"post_id" binding:"required"`
	Direction int8   `form:"direction,string" binding:"required,oneof=1 0 -1"`
}

// 获取帖子列表参数
type ParamPostList struct {
	CommunityID uint64 `json:"community_id" form:"community_id"`
	Page        int64  `json:"page" form:"page" binding:"min=1"`                   // 页码，最小为1
	Size        int64  `json:"size" form:"size" binding:"min=1,max=100"`           // 每页大小，1-100
	Order       string `json:"order" form:"order" binding:"oneof=time score view"` // 排序方式：time(时间), score(分数), view(访问量)
	Search      string `json:"search" form:"search"`
	UseIndex    bool   `json:"use_index" form:"use_index"` // 是否使用MySQL索引优化（默认true）
}

// 更新帖子参数
type UpdatePostForm struct {
	PostID      uint64 `json:"post_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	CommunityID uint64 `json:"community_id" binding:"required"`
}
