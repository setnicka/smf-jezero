package main

import (
	"fmt"
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
	Message     string
}

type teamIndexData struct {
	GeneralData

	RoundNumber    int
	GlobalState    int
	Money          int
	GameMessage    string
	SelectedAction int

	History []teamHistoryRecord
	Actions map[int]game.ActionDef
}

func teamIndexHandler(w http.ResponseWriter, r *http.Request) {
	data := teamIndexData{GeneralData: getGeneralData("Hra", r)}
	defer func() { executeTemplate(w, "teamIndex", data) }()

	team := server.state.GetTeam(getUser(r))

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.MessageType = "danger"
			data.Message = "Cannot parse form"
			return
		}

		if r.PostFormValue("setAction") != "" {
			server.state.CurrentActions[team.Login], _ = strconv.Atoi(r.PostFormValue("setAction"))
			server.state.Save()
			data.MessageType = "success"
			data.Message = fmt.Sprintf("Akce '%s' nastavena", game.GetActions()[server.state.CurrentActions[team.Login]].DisplayName)
		}
	}

	currentState := server.state.GetCurrentState()

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
