package apiutils

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleUser     Role = "user"
)

func (r Role) ValidRole() (Role, bool) {
	valid := false
	switch r {
	case RoleAdmin, RoleEmployee, RoleUser:
		valid = true
	}
	return r, valid
}
