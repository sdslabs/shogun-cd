package apiutils

type Role string

const (
	RoleAdmin Role = "admin"
)

func (r Role) ValidRole() (Role, bool) {
	valid := false
	switch r {
	case RoleAdmin:
		valid = true
	}
	return r, valid
}
