package room

import i "devzat/pkg/interfaces"

func (r *Room) GetAdmins() (map[string]string, error) {
	return r.Server().GetAdmins()
}

func (r *Room) IsAdmin(user i.User) bool {
	return r.Server().IsAdmin(user)
}
