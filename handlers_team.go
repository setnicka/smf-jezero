package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/setnicka/smf-jezero/game"
)

type teamHistoryRecord struct {
	RoundNumber int
	StartState  int
	StartMoney  int
	Action      int
	FinalState  int
	FinalMoney  int
	Message     template.HTML
}

type teamIndexData struct {
	GeneralData

	RoundNumber    int
	GlobalState    int
	Money          int
	GameMessage    template.HTML
	SelectedAction int

	History []teamHistoryRecord
	Actions map[int]game.ActionDef
}

func teamIndexHandler(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	currentState := server.state.GetCurrentState()
	var money int
	if currentStateTeam, found := currentState.Teams[team.Login]; found {
		money = currentStateTeam.Money
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse form"})
		}

		if r.PostFormValue("setAction") != "" {
			actionNumber, _ := strconv.Atoi(r.PostFormValue("setAction"))
			if action, found := game.GetActions()[actionNumber]; found {
				// Check if action could be performed
				if action.Check(currentState.GlobalState, money) {
					server.state.CurrentActions[team.Login] = actionNumber
					server.state.Save()
					setFlashMessage(w, r, FlashMessage{"success", fmt.Sprintf("Akce '%s' nastavena", action.DisplayName)})
				} else {
					setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Akce '%s' nemůže být nastavena, nejsou splněny podmínky", action.DisplayName)})
				}
			} else {
				setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Akce s indexem '%d' neexistuje", action)})
			}

		}
		http.Redirect(w, r, "", http.StatusSeeOther)
		return
	}

	data := teamIndexData{GeneralData: getGeneralData("Hra", w, r)}
	defer func() { executeTemplate(w, "teamIndex", data) }()

	data.RoundNumber = currentState.Number + 1
	data.GlobalState = currentState.GlobalState
	data.GameMessage = ""

	if currentStateTeam, found := currentState.Teams[team.Login]; found {
		data.Money = currentStateTeam.Money
		data.GameMessage = currentStateTeam.Message
	}

	// Construct history records
	for i := len(server.state.Rounds) - 1; i >= 1; i-- {
		currentRound := server.state.Rounds[i]
		lastRound := server.state.Rounds[i-1]

		record := teamHistoryRecord{
			RoundNumber: currentRound.Number,
			StartState:  lastRound.GlobalState,
			FinalState:  currentRound.GlobalState,
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

	data.Actions = game.GetActions()
	data.SelectedAction, _ = server.state.CurrentActions[team.Login]
}
