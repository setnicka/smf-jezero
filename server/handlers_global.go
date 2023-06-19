package server

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/setnicka/smf-jezero/game"
)

func (s *Server) getGeneralData(title string, w http.ResponseWriter, r *http.Request) GeneralData {
	data := GeneralData{
		Title:            title,
		CountdownActive:  !s.countdownTo.IsZero(),
		CountdownSeconds: int(math.Ceil(s.countdownTo.Sub(time.Now()).Seconds())),
	}
	if flashMessages := s.getFlashMessages(w, r); len(flashMessages) > 0 {
		data.MessageType = flashMessages[0].Type
		data.Message = flashMessages[0].Message
		log.Debugf("Flash message '%s'", flashMessages[0].Message)
	}
	return data
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.setFlashMessage(w, r, FlashMessage{"danger", "Cannot parse login form"})
		}
		login := r.PostFormValue("login")
		password := r.PostFormValue("password")
		if s.state.TeamCheckPassword(login, password) {
			session, _ := s.sessionStore.Get(r, sessionCookieName)
			session.Values["authenticated"] = true
			session.Values["user"] = login
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		s.setFlashMessage(w, r, FlashMessage{"danger", "Nesprávný login"})
		http.Redirect(w, r, "login", http.StatusSeeOther)
		return
	}

	data := s.getGeneralData("Přihlášení do hry", w, r)
	s.executeTemplate(w, "login", data)
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) calcGlobalHash() []string {
	currentState := s.state.GetLastState()

	return []string{
		strconv.Itoa(currentState.RoundNumber()),
		currentState.GlobalState.String(),
		strconv.FormatInt(s.countdownTo.Unix(), 10),
	}
}

func (s *Server) calcTeamHash(team *game.Team) string {
	hash := s.calcGlobalHash()

	actions := s.state.CurrentActions
	if action, found := actions[team.Login]; found {
		hash = append(hash, strconv.Itoa(action))
	}
	return strings.Join(hash, "-")
}

func (s *Server) calcOrgHash() string {
	hash := s.calcGlobalHash()

	actions := s.state.CurrentActions
	for _, team := range s.state.Teams {
		if action, found := actions[team.Login]; found {
			hash = append(hash, strconv.Itoa(action))
		}
	}
	return strings.Join(hash, "-")
}
