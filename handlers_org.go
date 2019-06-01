package main

import (
	"fmt"
	"html/template"
	"net/http"

	//"github.com/coreos/go-log/log"

	"github.com/setnicka/smf-jezero/game"
)

func orgLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse login form"})
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if login == ORG_LOGIN && password == ORG_PASSWORD {
			session, _ := server.sessionStore.Get(r, SESSION_COOKIE_NAME)
			session.Values["authenticated"] = true
			session.Values["org"] = true
			session.Save(r, w)
			http.Redirect(w, r, "dashboard", http.StatusSeeOther)
			return
		}
		setFlashMessage(w, r, FlashMessage{"danger", "Nesprávný login"})
		http.Redirect(w, r, "login", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Orgovský login", w, r) // Nothing special to add
	executeTemplate(w, "orgLogin", data)
}

////////////////////////////////////////////////////////////////////////////////

type orgTeamsData struct {
	GeneralData
	Teams map[string]game.Team
}

func orgTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse teams form"})
		}

		if r.PostFormValue("deleteTeam") != "" {
			if team := server.state.GetTeam(r.PostFormValue("deleteTeam")); team != nil {
				server.state.DeleteTeam(r.PostFormValue("deleteTeam"))
				setFlashMessage(w, r, FlashMessage{"success", "Team deleted"})
			}
		} else if r.PostFormValue("setPassword") != "" {
			if team := server.state.GetTeam(r.PostFormValue("login")); team != nil {
				server.state.TeamSetPassword(r.PostFormValue("login"), r.PostFormValue("setPassword"))
				setFlashMessage(w, r, FlashMessage{"success", "Password set"})
			}
		} else if r.PostFormValue("newTeamLogin") != "" {
			err := server.state.AddTeam(r.PostFormValue("newTeamLogin"), r.PostFormValue("newTeamName"))
			if err == nil {
				setFlashMessage(w, r, FlashMessage{"success", "Team added"})
			} else {

				setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Cannot add team due to error: %v", err)})
			}
		}
		http.Redirect(w, r, "teams", http.StatusSeeOther)
		return
	}

	data := orgTeamsData{GeneralData: getGeneralData("Týmy", w, r)}
	defer func() { executeTemplate(w, "orgTeams", data) }()

	data.Teams = map[string]game.Team{}
	for _, team := range server.state.GetTeams() {
		data.Teams[team.Login] = team
	}
}

////////////////////////////////////////////////////////////////////////////////

type orgDashboardData struct {
	GeneralData
	Teams          []string
	RoundNumber    int
	CurrentState   int
	CurrentActions []int
	History        []orgDashboardRoundRecord
	AllActions     map[int]game.ActionDef
}

type orgDashboardRoundRecord struct {
	RoundNumber int
	StartState  int
	FinalState  int
	Teams       []orgDashboardTeamRecord
}

type orgDashboardTeamRecord struct {
	Found      bool
	StartMoney int
	Action     int
	FinalMoney int
	Message    template.HTML
}

func orgDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse form"})
		}

		if r.PostFormValue("calculateRound") != "" {
			setFlashMessage(w, r, FlashMessage{"success", "Kolo spočítáno, výsledky níže"})
			server.state.EndRound()
		}

		if r.PostFormValue("resetGame") != "" {
			setFlashMessage(w, r, FlashMessage{"success", "Hra resetována"})
			server.state.InitGame()
		}

		if r.PostFormValue("sendState") != "" {
			if err := server.state.SendState(); err == nil {
				setFlashMessage(w, r, FlashMessage{"info", "Stav poslán"})
			} else {
				setFlashMessage(w, r, FlashMessage{"warning", fmt.Sprintf("Chyba při posílání stavu: %v", err)})
			}
		}

		http.Redirect(w, r, "dashboard", http.StatusSeeOther)
		return
	}

	data := orgDashboardData{GeneralData: getGeneralData("Org dashboard", w, r)}
	defer func() { executeTemplate(w, "orgDashboard", data) }()

	allTeams := server.state.GetTeams()

	data.AllActions = game.GetActions()
	data.RoundNumber = server.state.GetRoundNumber()
	data.CurrentState = server.state.GetLastState().GlobalState
	data.Teams = []string{}
	data.CurrentActions = []int{}
	for _, team := range allTeams {
		data.Teams = append(data.Teams, team.Name)
		data.CurrentActions = append(data.CurrentActions, server.state.CurrentActions[team.Login])
	}

	// Construct history records
	data.History = []orgDashboardRoundRecord{}
	for i := len(server.state.Rounds) - 1; i >= 1; i-- {
		currentRound := server.state.Rounds[i]
		lastRound := server.state.Rounds[i-1]

		record := orgDashboardRoundRecord{
			RoundNumber: currentRound.Number,
			StartState:  lastRound.GlobalState,
			FinalState:  currentRound.GlobalState,
			Teams:       []orgDashboardTeamRecord{},
		}

		for _, team := range allTeams {
			teamRecord := orgDashboardTeamRecord{}

			if teamState, found := currentRound.Teams[team.Login]; found {
				teamRecord.Found = true
				teamRecord.Action = teamState.Action
				teamRecord.FinalMoney = teamState.Money
				teamRecord.Message = teamState.Message
			}
			if lastTeamState, found := lastRound.Teams[team.Login]; found {
				teamRecord.StartMoney = lastTeamState.Money
			}

			record.Teams = append(record.Teams, teamRecord)
		}

		data.History = append(data.History, record)
	}

}
