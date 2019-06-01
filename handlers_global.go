package main

import (
	"encoding/gob"
	"net/http"

	"github.com/coreos/go-log/log"
)

func getGeneralData(title string, w http.ResponseWriter, r *http.Request) GeneralData {
	data := GeneralData{
		Title: title,
	}
	if flashMessages := getFlashMessages(w, r); len(flashMessages) > 0 {
		data.MessageType = flashMessages[0].Type
		data.Message = flashMessages[0].Message
		log.Debugf("Flash message '%s'", flashMessages[0].Message)
	}
	return data
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse login form"})
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if server.state.TeamCheckPassword(login, password) {
			session, _ := server.sessionStore.Get(r, SESSION_COOKIE_NAME)
			session.Values["authenticated"] = true
			session.Values["user"] = login
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		setFlashMessage(w, r, FlashMessage{"danger", "Nesprávný login"})
		http.Redirect(w, r, "login", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Přihlášení do hry", w, r)
	executeTemplate(w, "login", data)
}

////////////////////////////////////////////////////////////////////////////////

const FLASH_SESSION = "flash-session"

func setFlashMessage(w http.ResponseWriter, r *http.Request, message FlashMessage) {
	// Register the struct so encoding/gob knows about it
	gob.Register(FlashMessage{})

	session, err := server.sessionStore.Get(r, FLASH_SESSION)
	if err != nil {
		return
	}
	session.AddFlash(message)
	err = session.Save(r, w)
	if err != nil {
		log.Errorf("Cannot save flash message: %v", err)
	}
}

func getFlashMessages(w http.ResponseWriter, r *http.Request) []FlashMessage {
	// 1. Get session
	session, err := server.sessionStore.Get(r, FLASH_SESSION)
	if err != nil {
		return nil
	}

	// 2. Get flash messages
	parsedFlashes := []FlashMessage{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		for _, flash := range flashes {
			parsedFlashes = append(parsedFlashes, flash.(FlashMessage))
		}
	}

	// 3. Delete flash messages by saving session
	err = session.Save(r, w)
	if err != nil {
		log.Errorf("Problem during loading flash messages: %v", err)
	}

	return parsedFlashes
}
