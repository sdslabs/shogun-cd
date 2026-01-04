package dto

type HookInput struct {
	Pipeline  string `json:"pipeline" binding:"required,min=3,max=64,alphanum"`
	Alias     string `json:"alias" binding:"required,min=3,max=32"`
	Provider  string `json:"provider" binding:"required"`
	CreatedBy string `json:"-"`
}
