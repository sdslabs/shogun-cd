package dto

type HookInput struct {
	Pipeline  string `json:"pipeline" binding:"required"` // [TODO] finalise and add allowed formats for the pipeline name
	Alias     string `json:"alias" binding:"required,min=3,max=32"`
	Provider  string `json:"provider" binding:"required"`
	CreatedBy string `json:"-"`
}
