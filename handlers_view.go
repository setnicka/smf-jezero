package main

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

func getViewHash() string {
	return strings.Join(calcGlobalHash(), "-")
}

func getViewStatusData(w http.ResponseWriter, r *http.Request, currentState *game.RoundState) viewStatus {
	return viewStatus{
		GeneralData:   getGeneralData("Přehled", w, r),
		Hash:          getViewHash(),
		RoundNumber:   currentState.RoundNumber(),
		GlobalState:   currentState.GlobalState,
		GlobalMessage: currentState.GlobalMessage,
	}
}

func viewHashHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(getViewHash()))
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	currentState := server.state.GetLastState()

	executeTemplate(w, "viewIndex", getViewStatusData(w, r, currentState))
}

func viewStatusHandler(w http.ResponseWriter, r *http.Request) {
	currentState := server.state.GetLastState()

	render.JSON(w, r, getViewStatusData(w, r, currentState))
}
