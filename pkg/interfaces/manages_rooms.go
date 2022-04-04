package interfaces

type managesRooms interface {
	AllRooms() []Room
	AddRoom(Room)
	DeleteRoom(Room)
}
