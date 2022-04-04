package v2

import (
	i "devzat/pkg/interfaces"
	room "devzat/pkg/room/v2"
)

type roomManagement struct {
	mainRoom i.Room
	rooms    map[string]i.Room
}

func (rm *roomManagement) init(s *Server) error {
	rm.mainRoom = room.New("mainRoom")
	rm.mainRoom.SetServer(s)

	rm.rooms = make(map[string]i.Room)
	rm.rooms[rm.mainRoom.Name()] = rm.mainRoom

	return nil
}

func (s *Server) ChangeRoom(u i.User, roomName string) {
	if v, ok := s.rooms[roomName]; ok {
		u.SetRoom(v)
		return
	}

	r := room.New(roomName)
	s.rooms[roomName] = r

	r.SetServer(s)
	u.SetRoom(r)
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
