package server

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/setnicka/smf-jezero/game"
)

func (s *Server) getGeneralData(title string, w http.ResponseWriter, r *http.Request) GeneralData {
	data := GeneralData{
		Title:            title,
		CountdownActive:  !s.countdownTo.IsZero(),
		CountdownSeconds: int(math.Ceil(s.countdownTo.Sub(time.Now()).Seconds())),
		Variant:          s.variant,
	}
	if flashMessages := s.getFlashMessages(w, r); len(flashMessages) > 0 {
		data.MessageType = flashMessages[0].Type
		data.Message = flashMessages[0].Message
		slog.Debug("flash message", "message", flashMessages[0].Message)
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
		if s.state.TeamCheckLoginPassword(login, password) {
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

func (s *Server) getGlobalHash() []string {
	currentState := s.state.GetLastState()

	return []string{
		strconv.Itoa(currentState.RoundNumber()),
		currentState.GlobalState.Hash(),
		strconv.FormatInt(s.countdownTo.Unix(), 10),
	}
}

func (s *Server) calcGlobalHash() string {
	return strings.Join(s.getGlobalHash(), "-")
}

func (s *Server) calcTeamHash(team *game.Team) string {
	hash := s.getGlobalHash()

	actions := s.state.CurrentActions
	if action, found := actions[team.ID]; found {
		hash = append(hash, strconv.Itoa(int(action)))
	}
	return strings.Join(hash, "-")
}

func (s *Server) calcActionsHash() string {
	hash := []string{}
	actions := s.state.CurrentActions
	for _, team := range s.state.Teams {
		if action, found := actions[team.ID]; found {
			hash = append(hash, strconv.Itoa(int(action)))
		}
	}
	return strings.Join(hash, "-")
}
