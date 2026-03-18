package dto

type LoginInput struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

type UserFilter struct {
	Email    string
	IsActive *bool
}
