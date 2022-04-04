package server

import (
	i "devzat/pkg/interfaces"
	"devzat/pkg/room"
)

const (
	defaultMainRoomName = "lobby"
)

type roomManagement struct {
	mainRoom i.Room
	rooms    map[string]i.Room
}

func (rm *roomManagement) init(s *Server) error {
	s.AddRoom(room.New(defaultMainRoomName))
	rm.mainRoom = s.rooms[defaultMainRoomName]

	return nil
}

func (s *Server) ChangeRoom(u i.User, roomName string) {
	if existing, ok := s.rooms[roomName]; ok {
		u.SetRoom(existing)
		return
	}

	r := room.New(roomName)
	s.rooms[roomName] = r

	r.SetServer(s)
	u.SetRoom(r)

	if !s.serverSettings.Twitter.Offline {
		go s.SendCurrentUsersTwitterMessage()
	}
}

func (s *Server) MainRoom() i.Room {
	return s.mainRoom
}

func (s *Server) AllRooms() []i.Room {
	res := make([]i.Room, len(s.rooms))

	idx := 0
	for key := range s.rooms {
		res[idx] = s.rooms[key]
		idx++
	}

	return res
}

func (s *Server) AddRoom(room i.Room) {
	if s.rooms == nil {
		s.rooms = make(map[string]i.Room)
	}

	if s.mainRoom == nil {
		s.mainRoom = room
	}

	s.rooms[s.mainRoom.Name()] = s.mainRoom

	room.SetServer(s)

	if _, found := s.rooms[room.Name()]; !found {
		s.rooms[room.Name()] = room
	}
}

func (s *Server) DeleteRoom(room i.Room) {
	delete(s.rooms, room.Name())
}
