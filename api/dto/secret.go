package dto

type SecretInput struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type SecretFilter struct {
	Name string
}
