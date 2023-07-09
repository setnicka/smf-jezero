package server

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

func (s *Server) orgLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse login form"})
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if login == s.cfg.OrgLogin && password == s.cfg.OrgPassword {
			session, _ := s.sessionStore.Get(r, sessionCookieName)
			session.Values["authenticated"] = true
			session.Values["org"] = true
			session.Save(r, w)
			http.Redirect(w, r, "dashboard", http.StatusSeeOther)
			return
		}
		s.setFlashMessage(w, r, FlashMessage{"danger", "Nesprávný login"})
		http.Redirect(w, r, "login", http.StatusSeeOther)
		return
	}

	data := s.getGeneralData("Orgovský login", w, r) // Nothing special to add
	s.executeTemplate(w, "orgLogin", data)
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) orgHashHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s.calcOrgHash()))
}

type orgTeamsData struct {
	GeneralData
	Teams []game.Team
}

func (s *Server) orgTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse teams form"})
		}

		if r.PostFormValue("deleteTeam") != "" {
			if team := s.state.GetTeam(r.PostFormValue("deleteTeam")); team != nil {
				s.state.DeleteTeam(r.PostFormValue("deleteTeam"))
				s.setFlashMessage(w, r, FlashMessage{"success", "Team deleted"})
			}
		} else if r.PostFormValue("setPassword") != "" {
			if team := s.state.GetTeam(r.PostFormValue("login")); team != nil {
				s.state.TeamSetPassword(r.PostFormValue("login"), r.PostFormValue("setPassword"))
				s.setFlashMessage(w, r, FlashMessage{"success", "Password set"})
			}
		} else if r.PostFormValue("newTeamLogin") != "" {
			var part game.GamePart
			switch r.PostFormValue("newTeamPart") {
			case string(game.PartA):
				part = game.PartA
			case string(game.PartB):
				part = game.PartB
			default:
				s.setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Part '%s' is not valid game part", r.PostFormValue("newTeamPart"))})
				http.Redirect(w, r, "teams", http.StatusSeeOther)
				return
			}

			err := s.state.AddTeam(r.PostFormValue("newTeamLogin"), r.PostFormValue("newTeamName"), part)
			if err == nil {
				s.setFlashMessage(w, r, FlashMessage{"success", "Team added"})
			} else {

				s.setFlashMessage(w, r, FlashMessage{"danger", fmt.Sprintf("Cannot add team due to error: %v", err)})
			}
		}
		http.Redirect(w, r, "teams", http.StatusSeeOther)
		return
	}

	teams := s.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})
	s.executeTemplate(w, "orgTeams", orgTeamsData{
		GeneralData: s.getGeneralData("Týmy", w, r),
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
	Hash             string
	Teams            []game.Team
	RoundNumber      int
	CurrentState     game.GlobalState
	CurrentActions   []currentAction
	History          []orgDashboardRoundRecord
	AllActions       map[int]game.ActionDef
	NextCountdown    int
	HasNotifyTargets bool
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
func (s *Server) getHistoryRecords(teams []game.Team) []orgDashboardRoundRecord {
	history := []orgDashboardRoundRecord{}

	for i := len(s.state.Rounds) - 1; i >= 0; i-- {
		currentRound := s.state.Rounds[i]

		record := orgDashboardRoundRecord{
			RoundNumber: currentRound.Number,
			FinalState:  currentRound.GlobalState,
			Message:     currentRound.GlobalMessage,
			Teams:       []orgDashboardTeamRecord{},
		}

		lastRound := currentRound
		if i > 0 {
			lastRound = s.state.Rounds[i-1]
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

func (s *Server) orgDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse form"})
		}

		if r.PostFormValue("calculateRound") != "" {
			s.mutex.Lock()
			s.stopTimer()
			s.state.EndRound()
			s.mutex.Unlock()

			s.setFlashMessage(w, r, FlashMessage{"success", "Kolo spočítáno, výsledky níže"})
		}

		if r.PostFormValue("resetGame") != "" {
			s.mutex.Lock()
			s.stopTimer()
			s.state.InitGame()
			s.mutex.Unlock()

			s.setFlashMessage(w, r, FlashMessage{"success", "Hra resetována"})
		}

		if r.PostFormValue("sendState") != "" {
			if err := s.state.SendState(); err == nil {
				s.setFlashMessage(w, r, FlashMessage{"info", "Stav poslán"})
			} else {
				s.setFlashMessage(w, r, FlashMessage{"warning", fmt.Sprintf("Chyba při posílání stavu: %v", err)})
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
			s.mutex.Lock()
			s.countdownDuration = time.Duration(seconds) * time.Second
			s.nextCountdown = s.countdownDuration
			s.resetTimer()
			s.mutex.Unlock()

			s.setFlashMessage(w, r, FlashMessage{"success", fmt.Sprintf("Odpočet spuštěn, další kolo za %v", s.countdownDuration)})
		}
		if r.PostFormValue("submit-time-stop") != "" {
			s.mutex.Lock()
			s.stopTimer()
			s.mutex.Unlock()

			s.setFlashMessage(w, r, FlashMessage{"success", "Odpočet zastaven"})
		}
		if r.PostFormValue("submit-next-countdown") != "" && r.PostFormValue("countdown") != "" {
			seconds, err := strconv.Atoi(r.PostFormValue("countdown"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else if seconds < 0 {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			s.mutex.Lock()
			s.nextCountdown = time.Duration(seconds) * time.Second
			s.mutex.Unlock()

			s.setFlashMessage(w, r, FlashMessage{"success", fmt.Sprintf("Odpočet příštího kola nastaven na %v", s.nextCountdown)})
		}

		http.Redirect(w, r, "dashboard", http.StatusSeeOther)
		return
	}

	teams := s.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})

	data := orgDashboardData{
		GeneralData: s.getGeneralData("Stav hry", w, r),
		Hash:        s.calcOrgHash(),
		Teams:       teams,

		AllActions:     s.state.GetActions(),
		RoundNumber:    s.state.GetRoundNumber(),
		CurrentState:   s.state.GetLastState().GlobalState,
		CurrentActions: []currentAction{},
		History:        s.getHistoryRecords(teams),
		NextCountdown:  int(s.nextCountdown.Seconds()),

		HasNotifyTargets: s.state.HasNotifyTargets(),
	}
	for _, team := range teams {
		data.CurrentActions = append(data.CurrentActions, currentAction{
			Action: s.state.CurrentActions[team.Login],
			Team:   team,
		})
	}

	s.executeTemplate(w, "orgDashboard", data)
}

///////

type orgChartsData struct {
	GeneralData
	Teams   []game.Team
	History []orgDashboardRoundRecord
}

func (s *Server) orgChartsHandler(w http.ResponseWriter, r *http.Request) {
	teams := s.state.GetTeams()
	sort.Slice(teams, func(i, j int) bool {
		if teams[i].Part == teams[j].Part {
			return teams[i].Name < teams[j].Name
		}
		return teams[i].Part < teams[j].Part
	})

	s.executeTemplate(w, "orgCharts", orgChartsData{
		GeneralData: s.getGeneralData("Grafy", w, r),
		Teams:       teams,
		History:     s.getHistoryRecords(teams),
	})
}
