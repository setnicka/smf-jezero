package server

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/setnicka/smf-jezero/game"
)

type viewStatus struct {
	GeneralData
	Hash          string
	RoundNumber   int
	GlobalState   game.GlobalState
	GlobalMessage template.HTML
}

func (s *Server) getViewHash() string {
	return strings.Join(s.calcGlobalHash(), "-")
}

func (s *Server) getViewStatusData(w http.ResponseWriter, r *http.Request) viewStatus {
	currentState := s.state.GetLastState()

	return viewStatus{
		GeneralData:   s.getGeneralData("Přehled", w, r),
		Hash:          s.getViewHash(),
		RoundNumber:   currentState.RoundNumber(),
		GlobalState:   currentState.GlobalState,
		GlobalMessage: currentState.GlobalMessage,
	}
}

func (s *Server) viewHashHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(s.getViewHash()))
}

func (s *Server) viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	s.executeTemplate(w, s.variant.ViewTemplateName(), s.getViewStatusData(w, r))
}

func (s *Server) viewStatusHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, s.getViewStatusData(w, r))
}
