package server

import (
	"devzat/pkg/interfaces"
	"devzat/pkg/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultBansSize = 10
	tooManyLogins   = 6
)

const (
	defaultBansFile               = "bans.json"
	fmtDefaultBannedLoginResponse = `
		**You are banned**. 
		If you feel this was a mistake, please reach out at github.com/quackduck/devzat/issues 
		or email igoel.mail@gmail.com. 
	
		Please include the following information: [ID %v]
	`
)

type banManagement struct {
	bans []models.Ban
}

func (bm *banManagement) init(s *Server) error {
	bm.bans = make([]models.Ban, 0, defaultBansSize)
	if err := s.ReadBans(); err != nil {
		return fmt.Errorf("could not read bans file: %s", err)
	}

	return nil
}

func (bm *banManagement) GetBanList() []models.Ban {
	dupe := make([]models.Ban, 0)
	return append(dupe, bm.bans...)
}

func (s *Server) SaveBans() error {
	f, err := os.Create(defaultBansFile)
	if err != nil {
		s.Error().Err(err)
		return fmt.Errorf("could not create banList file: %v", err)
	}

	j := json.NewEncoder(f)
	j.SetIndent("", "   ")

	if err = j.Encode(defaultBansSize); err != nil {
		s.MainRoom().BotCast("error saving banList: " + err.Error())
		s.Error().Err(err)

		return err
	}

	return f.Close()
}

func (s *Server) ReadBans() error {
	fp := filepath.Join(s.ConfigDir(), defaultBansFile)

	f, err := os.Open(fp)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		if errCreate := s.makeDefaultBansFile(fp); errCreate != nil {
			return errCreate
		}
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if errJson := json.Unmarshal(data, &s.bans); errJson != nil {
		msg := fmt.Sprintf("error decoding json: %v", errJson)

		s.Info().Msg(msg)

		return errJson
	}

	return f.Close()
}

func (s *Server) makeDefaultBansFile(path string) error {
	f, errCreate := os.Create(path)
	if errCreate != nil {
		return errCreate
	}

	_, _ = f.WriteString("[]")

	return nil
}

// BansContains reports if the addr or id is found in the bans list
func (s *Server) BansContains(addr string, id string) bool {
	for i := 0; i < len(s.bans); i++ {
		if s.bans[i].Addr == addr || s.bans[i].ID == id {
			return true
		}
	}

	return false
}

func (s *Server) BanUser(strBanner string, victim interfaces.User) {
	s.bans = append(s.bans, models.Ban{victim.Addr(), victim.ID()})
	_ = s.SaveBans()
	victim.Close(victim.Name() + " has been banned by " + strBanner)
}

func (s *Server) BanUserForDuration(banner string, victim interfaces.User, dur time.Duration) {
	banner = fmt.Sprintf("%s for %s", banner, dur.String())
	msg := fmt.Sprintf("%s has been banned by %s", victim.Name(), banner)

	s.BanUser(msg, victim)

	go func(id string) {
		time.Sleep(dur)
		_ = s.UnbanUser(victim.Name())
	}(victim.ID()) // evaluate id now, call unban with that value later
}

// we just use a map for easy lookup
type banList = map[string]interface{}

func (s *Server) UnbanUser(toUnban string) error {
	list := s.getBans()
	if _, found := list[toUnban]; !found {
		return nil
	}

	delete(list, toUnban)

	if _, found := list[toUnban]; found {
		// when generating the list, we create an entry for both the addr and id
		// so, if it had both, we delete the corresponding entry here
		delete(list, toUnban)
	}

	if err := s.SaveBans(); err != nil {
		return fmt.Errorf("could not save bans: %v", err)
	}

	return nil
}

func (s *Server) getBans() banList {
	bl := make(banList)

	for _, b := range s.bans {
		if b.ID != "" {
			bl[b.ID] = b.Addr // hack, used for reverse lookup
		}

		if b.Addr != "" {
			bl[b.Addr] = b.ID // hack, used for reverse lookup
		}
	}

	return bl
}
