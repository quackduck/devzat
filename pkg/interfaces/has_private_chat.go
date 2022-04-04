package interfaces

type hasPrivateChat interface {
	DMTarget() User
	SetDMTarget(User)
}
