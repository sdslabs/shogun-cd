package dto

type SecretInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"-"`
}

type SecretFilter struct {
	Name string
}
