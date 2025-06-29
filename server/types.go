package server

import "github.com/setnicka/smf-jezero/game"

// Cookies
const (
	sessionCookieName = "jezero-session"
	flashSessionName  = "jezero-flash"
)

// GeneralData for rendering page
type GeneralData struct {
	BaseURL          string
	Title            string
	User             string
	MessageType      string
	Message          string
	CountdownActive  bool
	CountdownSeconds int
	Variant          game.Variant
}

// FlashMessage holds type and content of flash message displayed to the user
type FlashMessage struct {
	Type    string
	Message string
}
