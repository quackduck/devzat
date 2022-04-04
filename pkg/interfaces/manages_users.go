package interfaces

type managesUsers interface {
	ChangeRoom(u User, roomName string)
	Disconnect(User)
	Kick(u User, reason string)
	AllUsers() []User
	UserDuplicate(a string) (User, bool)
	FindUserByName(name string) (User, bool)
	PrintUsers() string
}
