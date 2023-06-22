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
	"github.com/setnicka/smf-jezero/config"
)

// Init the Game
func Init(cfg config.GameConfig) *State {
	log.Debug("Initializing game state")
	state := &State{
		cfg:     cfg,
		Teams:   []Team{},
	}

	// Try to load previously saved state
	jsonFile, err := os.Open(cfg.StateFile)
	if err == nil {
		defer jsonFile.Close()

		log.Debugf("Loading state from file '%s'", cfg.StateFile)
		jsonBytes, _ := ioutil.ReadAll(jsonFile)
		if err = json.Unmarshal(jsonBytes, &state); err != nil {
			log.Errorf("Problem during loading state from file: %v", err)
		} else {
			log.Debug("State loaded")
		}
	}

	if len(state.Rounds) == 0 {
		state.InitGame()
	}

	return state
}

func (s *State) GetTeams() []Team {
	return s.Teams
}

func (s *State) GetTeam(login string) *Team {
	for i, team := range s.Teams {
		if team.Login == login {
			return &s.Teams[i]
		}
	}
	return nil
}

func (s *State) AddTeam(login string, name string, part GamePart) error {
	if s.GetTeam(login) != nil {
		return fmt.Errorf("Team with name '%s' already exists", login)
	}
	s.Teams = append(s.Teams, Team{Login: login, Name: name, Part: part})
	s.Save()
	return nil
}

func (s *State) DeleteTeam(login string) error {
	for i, team := range s.Teams {
		if team.Login == login {
			s.Teams = append(s.Teams[:i], s.Teams[i+1:]...)
			s.Save()
			return nil
		}
	}
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
	if _, err := os.Stat(s.cfg.StateFile); err == nil {
		// Ensure dir exists
		os.MkdirAll(s.cfg.BackupDir, os.ModePerm)
		os.Rename(s.cfg.StateFile, path.Join(s.cfg.BackupDir, fmt.Sprintf("%s%s", s.cfg.StateFile, time.Now().Format(".150405.00")))) // 2006-01-02_150405
	}

	// 2. Marshal state into json
	bytes, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Errorf("Cannot save actual state into json: %v", err)
		return
	}

	if err := ioutil.WriteFile(s.cfg.StateFile, bytes, 0644); err != nil {
		log.Errorf("Cannot save json of actual state into file '%s': %v", s.cfg.StateFile, err)
	}
}
