package main

import (
	"net/http"
)

func getGeneralData(title string, r *http.Request) GeneralData {
	return GeneralData{
		Title: title,
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := getGeneralData("Přihlášení do hry", r)
	defer func() {
		executeTemplate(w, "login", data)
	}()

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			data.MessageType = "danger"
			data.Message = "Cannot parse login form"
			return
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if server.state.TeamCheckPassword(login, password) {
			session, _ := server.sessionStore.Get(r, SESSION_COOKIE_NAME)
			session.Values["authenticated"] = true
			session.Values["user"] = login
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			data.MessageType = "danger"
			data.Message = "Nesprávný login"
		}
	}
}
