package interfaces

type hasAdmins interface {
	GetAdmins() (map[string]string, error)
	IsAdmin(User) (bool, error)
}

type hasPrivilege interface {
	IsAdmin() bool
}
