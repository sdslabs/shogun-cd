package users

type CreateParams struct {
	Email    string
	Password string
}

type FilterParams struct {
	Email    string
	IsActive *bool
}
