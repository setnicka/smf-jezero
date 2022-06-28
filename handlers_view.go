package main

import (
	"html/template"
	"net/http"

	"github.com/setnicka/smf-jezero/game"
)

type viewData struct {
	GeneralData

	GlobalState   game.GlobalState
	GlobalMessage template.HTML
}

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	currentState := server.state.GetLastState()

	data := viewData{
		GeneralData:   getGeneralData("Přehled", w, r),
		GlobalState:   currentState.GlobalState,
		GlobalMessage: currentState.GlobalMessage,
	}

	executeTemplate(w, "viewIndex", data)
}
