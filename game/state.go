// Package game contains the whole game mechanic and saving/loading of states.
package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/setnicka/smf-jezero/config"
)

// Init the Game
func Init(cfg config.GameConfig, variant Variant) *State {
	slog.Debug("initializing game state")
	state := &State{
		cfg:     cfg,
		variant: variant,
		Teams:   []Team{},
	}

	// Try to load previously saved state
	jsonFile, err := os.Open(cfg.StateFile)
	if err == nil {
		defer jsonFile.Close()
		slog := slog.With("file", cfg.StateFile)

		slog.Debug("loading state")
		jsonBytes, _ := io.ReadAll(jsonFile)
		if err = json.Unmarshal(jsonBytes, &state); err != nil {
			slog.Error("problem during loading state from file", "err", err)
		} else {
			slog.Debug("state loaded")
		}
	}

	if len(state.Rounds) == 0 {
		state.InitGame()
	}

	return state
}

// HasNotifyTargets return true if the tcp_notify section of the config is nonempty
func (s *State) HasNotifyTargets() bool {
	return len(s.cfg.TCPNotify) > 0
}

// GetTeams returns config of all teams of the given game
func (s *State) GetTeams() []Team {
	return s.Teams
}

// GetTeamByLogin identified by the login
func (s *State) GetTeamByLogin(login string) *Team {
	for i, team := range s.Teams {
		if team.Login == login {
			return &s.Teams[i]
		}
	}
	return nil
}

// GetTeam identified by the ID
func (s *State) GetTeam(teamID TeamID) *Team {
	for i, team := range s.Teams {
		if team.ID == teamID {
			return &s.Teams[i]
		}
	}
	return nil
}

// AddTeam adds team with given parameters (notice: password is not set)
func (s *State) AddTeam(login string, name string, part PartID) error {
	if s.GetTeamByLogin(login) != nil {
		return fmt.Errorf("Team with name '%s' already exists", login)
	}
	s.Teams = append(s.Teams, Team{ID: TeamID(login), Login: login, Name: name, Part: part})
	s.Save()
	return nil
}

// DeleteTeam identified by the ID
func (s *State) DeleteTeam(teamID TeamID) error {
	for i, team := range s.Teams {
		if team.ID == teamID {
			s.Teams = append(s.Teams[:i], s.Teams[i+1:]...)
			s.Save()
			return nil
		}
	}
	return fmt.Errorf("Cannot find team with ID '%s'", teamID)
}

// TeamSetPassword sets salted password of the team
func (s *State) TeamSetPassword(teamID TeamID, password string) {
	team := s.GetTeam(teamID)
	if team == nil {
		return
	}
	slog.Debug("new password set", "team", teamID)
	team.Salt, _ = genRandomString(12)
	team.Password = fmt.Sprintf("%x", sha256.Sum256([]byte(team.Salt+password)))
	s.Save()

}

// TeamCheckLoginPassword returns true if the password matches for team with given login
func (s *State) TeamCheckLoginPassword(login string, password string) bool {
	team := s.GetTeamByLogin(login)
	if team == nil {
		return false
	}
	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(team.Salt+password)))
	return (pass == team.Password)
}

// Save the game state to the file specified in configuration
func (s *State) Save() {
	slog := slog.With("file", s.cfg.StateFile)

	// 1. If exists current state move it into folder
	if _, err := os.Stat(s.cfg.StateFile); err == nil {
		// Ensure dir exists
		os.MkdirAll(s.cfg.BackupDir, os.ModePerm)
		os.Rename(s.cfg.StateFile, path.Join(s.cfg.BackupDir, fmt.Sprintf("%s%s", s.cfg.StateFile, time.Now().Format(".150405.00")))) // 2006-01-02_150405
	}

	// 2. Marshal state into json
	bytes, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		slog.Error("cannot marshall actual state into json", "err", err)
		return
	}

	if err := os.WriteFile(s.cfg.StateFile, bytes, 0644); err != nil {
		slog.Error("cannot save state", "err", err)
		return
	}

	slog.Debug("state saved")
}
