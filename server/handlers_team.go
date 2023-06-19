package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/setnicka/smf-jezero/game"
)

func (s *Server) getRoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", s.state.GetRoundNumber())
}

type teamHistoryRecord struct {
	RoundNumber   int
	StartState    game.GlobalState
	StartMoney    int
	Action        int
	FinalState    game.GlobalState
	FinalMoney    int
	Message       template.HTML
	GlobalMessage template.HTML
}

type teamIndexData struct {
	GeneralData

	Hash string

	Team           *game.Team
	RoundNumber    int
	GlobalState    game.GlobalState
	GlobalMessage  template.HTML
	Money          int
	GameMessage    template.HTML
	SelectedAction int

	History []teamHistoryRecord
	Actions map[int]game.ActionDef
}

func (s *Server) teamHashHandler(w http.ResponseWriter, r *http.Request) {
	team := s.state.GetTeam(s.getUser(r))
	w.Write([]byte(s.calcTeamHash(team)))
}

func (s *Server) teamIndexHandler(w http.ResponseWriter, r *http.Request) {
	team := s.state.GetTeam(s.getUser(r))
	currentState := s.state.GetLastState()
	var money int
	if currentStateTeam, found := currentState.Teams[team.Login]; found {
		money = currentStateTeam.Money
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse form"})
		}

		if r.PostFormValue("setAction") != "" {
			actionNumber, _ := strconv.Atoi(r.PostFormValue("setAction"))
			if action, found := game.GetActions()[actionNumber]; found {
				// Check if action could be performed
				if action.Check(currentState.GlobalState[team.Part], money) {
					s.state.CurrentActions[team.Login] = actionNumber
					s.state.Save()
					s.setFlashMessage(w, r, FlashMessage{"success", fmt.Sprintf("Akce '%s' nastavena", action.DisplayName)})
				} else {
					s.setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Akce '%s' nemůže být nastavena, nejsou splněny podmínky", action.DisplayName)})
				}
			} else {
				s.setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Akce s indexem '%d' neexistuje", actionNumber)})
			}

		}
		http.Redirect(w, r, "", http.StatusSeeOther)
		return
	}

	data := teamIndexData{
		GeneralData:   s.getGeneralData("Hra", w, r),
		Hash:          s.calcTeamHash(team),
		Team:          team,
		RoundNumber:   currentState.RoundNumber(),
		GlobalState:   currentState.GlobalState,
		GlobalMessage: currentState.GlobalMessage,
		Actions:       game.GetActions(),
	}
	if currentStateTeam, found := currentState.Teams[team.Login]; found {
		data.Money = currentStateTeam.Money
		data.GameMessage = currentStateTeam.Message
	}
	data.SelectedAction, _ = s.state.CurrentActions[team.Login]

	// Construct history records
	for i := len(s.state.Rounds) - 1; i >= 1; i-- {
		currentRound := s.state.Rounds[i]
		lastRound := s.state.Rounds[i-1]

		record := teamHistoryRecord{
			RoundNumber:   currentRound.Number,
			StartState:    lastRound.GlobalState,
			FinalState:    currentRound.GlobalState,
			GlobalMessage: currentRound.GlobalMessage,
		}

		if teamState, found := currentRound.Teams[team.Login]; found {
			record.Action = teamState.Action
			record.FinalMoney = teamState.Money
			record.Message = teamState.Message
		}
		if lastTeamState, found := lastRound.Teams[team.Login]; found {
			record.StartMoney = lastTeamState.Money
		}

		data.History = append(data.History, record)
	}

	s.executeTemplate(w, "teamIndex", data)
}
