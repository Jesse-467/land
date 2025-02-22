package models

const (
	OrderTime  = "time"
	OrderScore = "score"
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
	Page        int64  `json:"page" form:"page"`
	Size        int64  `json:"size" form:"size"`
	Order       string `json:"order" form:"order" example:"score"`
	Search      string `json:"search" form:"search"`
}
