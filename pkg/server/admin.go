package server

import (
	i "devzat/pkg/interfaces"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	defaultAdminsFileName = "admins.json"
)

type (
	AdminID      = string
	AdminInfo    = string
	AdminInfoMap = map[AdminID]AdminInfo
)

func (s *Server) GetAdmins() (AdminInfoMap, error) {
	if _, err := os.Stat(defaultAdminsFileName); err == os.ErrNotExist {
		return nil, errors.New("make an admins.json file to add admins")
	}

	data, err := ioutil.ReadFile("admins.json")
	if err != nil {
		return nil, fmt.Errorf("error reading admins.json: %s", err)
	}

	adminsList := make(AdminInfoMap)

	if err = json.Unmarshal(data, &adminsList); err != nil {
		return nil, fmt.Errorf("bad json: %v", err)
	}

	return adminsList, nil
}

func (s *Server) GiveAdmin(user i.User) error {
	if s.adminsMap == nil {
		s.adminsMap = make(adminsMap)
	}

	if _, alreadyAdmin := s.adminsMap[user.ID()]; alreadyAdmin {
		return fmt.Errorf("user '%s' is already an admin", user.Name())
	}

	s.adminsMap[user.ID()] = nil

	return nil
}

func (s *Server) RevokeAdmin(user i.User) error {
	if s.adminsMap == nil {
		s.adminsMap = make(adminsMap)
	}

	if _, isAdmin := s.adminsMap[user.ID()]; !isAdmin {
		return fmt.Errorf("user '%s' is not an admin", user.Name())
	}

	delete(s.adminsMap, user.ID())

	return nil
}

func (s *Server) IsAdmin(user i.User) bool {
	_, found := s.adminsMap[user.Name()]
	return found
}
