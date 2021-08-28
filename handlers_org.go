package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

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

func orgHashHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(calcOrgHash()))
}

type orgTeamsData struct {
	GeneralData
	Teams []game.Team
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
			var part game.GamePart
			switch r.PostFormValue("newTeamPart") {
			case string(game.PartA):
				part = game.PartA
			case string(game.PartB):
				part = game.PartB
			default:
				setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Part '%s' is not valid game part", r.PostFormValue("newTeamPart"))})
				http.Redirect(w, r, "teams", http.StatusSeeOther)
				return
			}

			err := server.state.AddTeam(r.PostFormValue("newTeamLogin"), r.PostFormValue("newTeamName"), part)
			if err == nil {
				setFlashMessage(w, r, FlashMessage{"success", "Team added"})
			} else {

				setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Cannot add team due to error: %v", err)})
			}
		}
		http.Redirect(w, r, "teams", http.StatusSeeOther)
		return
	}

	teams := server.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})
	executeTemplate(w, "orgTeams", orgTeamsData{
		GeneralData: getGeneralData("Týmy", w, r),
		Teams:       teams,
	})
}

////////////////////////////////////////////////////////////////////////////////

type currentAction struct {
	Action int
	Team   game.Team
}

type orgDashboardData struct {
	GeneralData
	Hash           string
	Teams          []game.Team
	RoundNumber    int
	CurrentState   game.GlobalState
	CurrentActions []currentAction
	History        []orgDashboardRoundRecord
	AllActions     map[int]game.ActionDef
}

type orgDashboardRoundRecord struct {
	RoundNumber int
	StartState  game.GlobalState
	FinalState  game.GlobalState
	Message     template.HTML
	Teams       []orgDashboardTeamRecord
}

type orgDashboardTeamRecord struct {
	Team       game.Team
	Found      bool
	StartMoney int
	Action     int
	FinalMoney int
	Message    template.HTML
}

// Construct history records for teams in given order
func getHistoryRecords(teams []game.Team) []orgDashboardRoundRecord {
	history := []orgDashboardRoundRecord{}

	for i := len(server.state.Rounds) - 1; i >= 0; i-- {
		currentRound := server.state.Rounds[i]

		record := orgDashboardRoundRecord{
			RoundNumber: currentRound.Number,
			FinalState:  currentRound.GlobalState,
			Message:     currentRound.GlobalMessage,
			Teams:       []orgDashboardTeamRecord{},
		}

		lastRound := currentRound
		if i > 0 {
			lastRound = server.state.Rounds[i-1]
			record.StartState = lastRound.GlobalState
		}

		for _, team := range teams {
			teamRecord := orgDashboardTeamRecord{
				Team: team,
			}

			if teamState, found := currentRound.Teams[team.Login]; found {
				teamRecord.Found = true
				teamRecord.Action = teamState.Action
				teamRecord.FinalMoney = teamState.Money
				teamRecord.Message = teamState.Message
			}
			if i > 0 {
				if lastTeamState, found := lastRound.Teams[team.Login]; found {
					teamRecord.StartMoney = lastTeamState.Money
				}
			}

			record.Teams = append(record.Teams, teamRecord)
		}

		history = append(history, record)
	}
	return history
}

func orgDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse form"})
		}

		if r.PostFormValue("calculateRound") != "" {
			server.mutex.Lock()
			server.stopTimer()
			server.state.EndRound()
			server.mutex.Unlock()

			setFlashMessage(w, r, FlashMessage{"success", "Kolo spočítáno, výsledky níže"})
		}

		if r.PostFormValue("resetGame") != "" {
			server.mutex.Lock()
			server.stopTimer()
			server.state.InitGame()
			server.mutex.Unlock()

			setFlashMessage(w, r, FlashMessage{"success", "Hra resetována"})
		}

		if r.PostFormValue("sendState") != "" {
			if err := server.state.SendState(); err == nil {
				setFlashMessage(w, r, FlashMessage{"info", "Stav poslán"})
			} else {
				setFlashMessage(w, r, FlashMessage{"warning", fmt.Sprintf("Chyba při posílání stavu: %v", err)})
			}
		}

		if r.PostFormValue("submit-time-start") != "" && r.PostFormValue("countdown") != "" {
			seconds, err := strconv.Atoi(r.PostFormValue("countdown"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else if seconds <= 0 {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			server.mutex.Lock()
			server.countdownDuration = time.Duration(seconds) * time.Second
			server.resetTimer()
			server.mutex.Unlock()

			setFlashMessage(w, r, FlashMessage{"success", fmt.Sprintf("Odpočet spuštěn, další kolo za %v", server.countdownDuration)})
		}
		if r.PostFormValue("submit-time-stop") != "" {
			server.mutex.Lock()
			server.stopTimer()
			server.mutex.Unlock()

			setFlashMessage(w, r, FlashMessage{"success", "Odpočet zastaven"})
		}

		http.Redirect(w, r, "dashboard", http.StatusSeeOther)
		return
	}

	teams := server.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})

	data := orgDashboardData{
		GeneralData: getGeneralData("Stav hry", w, r),
		Hash:        calcOrgHash(),
		Teams:       teams,

		AllActions:     game.GetActions(),
		RoundNumber:    server.state.GetRoundNumber(),
		CurrentState:   server.state.GetLastState().GlobalState,
		CurrentActions: []currentAction{},
		History:        getHistoryRecords(teams),
	}
	for _, team := range teams {
		data.CurrentActions = append(data.CurrentActions, currentAction{
			Action: server.state.CurrentActions[team.Login],
			Team:   team,
		})
	}

	executeTemplate(w, "orgDashboard", data)
}

///////

type orgChartsData struct {
	GeneralData
	Teams   []game.Team
	History []orgDashboardRoundRecord
}

func orgChartsHandler(w http.ResponseWriter, r *http.Request) {
	teams := server.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})

	executeTemplate(w, "orgCharts", orgChartsData{
		GeneralData: getGeneralData("Grafy", w, r),
		Teams:       teams,
		History:     getHistoryRecords(teams),
	})
}
