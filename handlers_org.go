package main

import (
	"fmt"
	"net/http"
	//"github.com/coreos/go-log/log"
)

func orgLoginHandler(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("Orgovský login", r) // Nothing special to add
	defer func() { executeTemplate(w, "orgLogin", data) }()

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.MessageType = "danger"
			data.Message = "Cannot parse login form"
			return
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if login == ORG_LOGIN && password == ORG_PASSWORD {
			session, _ := server.sessionStore.Get(r, SESSION_COOKIE_NAME)
			session.Values["authenticated"] = true
			session.Values["org"] = true
			session.Save(r, w)
			http.Redirect(w, r, "dashboard", http.StatusSeeOther)
		} else {
			data.MessageType = "info"
			data.Message = "Nepsrávný login"
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type orgTeamsData struct {
	GeneralData
	Teams map[string]string
}

func orgTeamsHandler(w http.ResponseWriter, r *http.Request) {
	data := orgTeamsData{GeneralData: getGeneralData("Týmy", r)}
	defer func() { executeTemplate(w, "orgTeams", data) }()

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.MessageType = "danger"
			data.Message = "Cannot parse teams form"
			return
		}

		if r.PostFormValue("deleteTeam") != "" {
			if team := server.state.GetTeam(r.PostFormValue("deleteTeam")); team != nil {
				server.state.DeleteTeam(r.PostFormValue("deleteTeam"))
				data.MessageType = "success"
				data.Message = "Team deleted"
			}
		} else if r.PostFormValue("setPassword") != "" {
			if team := server.state.GetTeam(r.PostFormValue("login")); team != nil {
				server.state.TeamSetPassword(r.PostFormValue("login"), r.PostFormValue("setPassword"))
				data.MessageType = "success"
				data.Message = "Password set"
			}
		} else if r.PostFormValue("newTeamLogin") != "" {
			err := server.state.AddTeam(r.PostFormValue("newTeamLogin"), r.PostFormValue("newTeamName"))
			if err == nil {
				data.MessageType = "success"
				data.Message = "Team added"
			} else {
				data.MessageType = "danger"
				data.Message = fmt.Sprintf("Cannot add team due to error: %v", err)
			}
		}
	}

	data.Teams = map[string]string{}
	for _, team := range server.state.GetTeams() {
		data.Teams[team.Login] = team.Name
	}
}

////////////////////////////////////////////////////////////////////////////////

type teamServiceResult struct {
	Completed     bool
	CompletedTime string
	Tries         int
}

type teamResult struct {
	Name    string
	Results []teamServiceResult
}

type orgDashboardData struct {
	GeneralData
	SecretServices []string
	Teams          []teamResult
}

func orgDashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := orgDashboardData{GeneralData: getGeneralData("Výsledky", r)}
	defer func() { executeTemplate(w, "orgDashboard", data) }()
}
