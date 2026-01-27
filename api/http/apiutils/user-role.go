package apiutils

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) ValidRole() (Role, bool) {
	valid := false
	switch r {
	case RoleAdmin, RoleUser:
		valid = true
	}
	return r, valid
}
