package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/coreos/go-log/log"
)

const (
	LOGS_DIR       = "logs"
	STATE_FILENAME = "state.json"
)

func Init() *State {
	log.Debug("Initializing game state")
	state := &State{
		Teams: []team{},
	}

	// Try to load previously saved state
	jsonFile, err := os.Open(STATE_FILENAME)
	if err == nil {
		defer jsonFile.Close()

		log.Debugf("Loading state from file '%s'", STATE_FILENAME)
		jsonBytes, _ := ioutil.ReadAll(jsonFile)
		if err = json.Unmarshal(jsonBytes, &state); err != nil {
			log.Errorf("Problem during loading state from file: %v", err)
		} else {
			log.Debug("State loaded")
		}
	}

	return state
}

func (s *State) GetTeams() []team {
	return s.Teams
}

func (s *State) GetTeam(login string) *team {
	for i, team := range s.Teams {
		if team.Login == login {
			return &s.Teams[i]
		}
	}
	return nil
}

func (s *State) AddTeam(login string, name string) error {
	if s.GetTeam(login) != nil {
		return fmt.Errorf("Team with name '%s' already exists", login)
	}
	s.Teams = append(s.Teams, team{Login: login, Name: name})
	s.Save()
	return nil
}

func (s *State) DeleteTeam(login string) error {
	for i, team := range s.Teams {
		if team.Login == login {
			s.Teams = append(s.Teams[:i], s.Teams[i+1:]...)
			return nil
		}
	}
	s.Save()
	return fmt.Errorf("Cannot find team with login '%s'", login)
}

func (s *State) TeamSetPassword(login string, password string) {
	team := s.GetTeam(login)
	if team == nil {
		return
	}
	log.Debugf("Saving new password for team '%s'", login)
	team.Salt, _ = genRandomString(12)
	team.Password = fmt.Sprintf("%x", sha256.Sum256([]byte(team.Salt+password)))
	s.Save()

}

func (s *State) TeamCheckPassword(login string, password string) bool {
	team := s.GetTeam(login)
	if team == nil {
		return false
	}
	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(team.Salt+password)))
	return (pass == team.Password)
}

func (s *State) Save() {
	log.Debug("Saving actual state into file")
	// 1. If exists current state move it into folder
	if _, err := os.Stat(STATE_FILENAME); err == nil {
		// Ensure dir exists
		os.MkdirAll(LOGS_DIR, os.ModePerm)
		os.Rename(STATE_FILENAME, path.Join(LOGS_DIR, fmt.Sprintf("%s%s", STATE_FILENAME, time.Now().Format(".150405.00")))) // 2006-01-02_150405
	}

	// 2. Marshal state into json
	bytes, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Errorf("Cannot save actual state into json: %v", err)
		return
	}

	if err := ioutil.WriteFile(STATE_FILENAME, bytes, 0644); err != nil {
		log.Errorf("Cannot save json of actual state into file '%s': %v", STATE_FILENAME, err)
	}
}
