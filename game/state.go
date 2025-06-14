// Package game contains the whole game mechanic and saving/loading of states.
package game

import (
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
func (s *State) AddTeam(login string, part PartID) error {
	newTeams := []Team{}

	if part == All {
		for _, p := range allParts {
			partLogin := login + string(p)
			newTeams = append(newTeams, Team{
				ID:       TeamID(partLogin),
				BaseID:   TeamID(login),
				Login:    partLogin,
				Name:     partLogin,
				Part:     p,
				Password: genRandomPassword(6),
			})
		}
	} else {
		newTeams = append(newTeams, Team{
			ID:       TeamID(login),
			BaseID:   TeamID(login),
			Login:    login,
			Name:     login,
			Part:     part,
			Password: genRandomPassword(6),
		})
	}

	for _, team := range newTeams {
		if s.GetTeam(team.ID) != nil || s.GetTeamByLogin(team.Login) != nil {
			return fmt.Errorf("Team with ID or login '%s' already exists", login)
		}
	}

	s.Teams = append(s.Teams, newTeams...)
	s.Save()
	return nil
}

// DeleteTeamsByBase deletes all teams with given baseID
func (s *State) DeleteTeamsByBase(baseID TeamID) {
	newTeams := []Team{}
	for _, team := range s.Teams {
		if team.BaseID != baseID {
			newTeams = append(newTeams, team)
		}
	}
	s.Teams = newTeams
	s.Save()
}

// SetTeamName sets name to all teams with given baseID
func (s *State) SetTeamName(baseID TeamID, name string) {
	for i := range s.Teams {
		if s.Teams[i].BaseID == baseID {
			s.Teams[i].Name = name
		}
	}
	s.Save()
}

// TeamCheckLoginPassword returns true if the password matches for team with given login
func (s *State) TeamCheckLoginPassword(login string, password string) bool {
	team := s.GetTeamByLogin(login)
	if team == nil {
		return false
	}
	return (password == team.Password)
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
