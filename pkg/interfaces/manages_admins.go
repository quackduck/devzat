package interfaces

type managesAdmins interface {
	GiveAdmin(User) error
	RevokeAdmin(User) error
}
