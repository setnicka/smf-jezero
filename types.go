package main

// GeneralData for rendering page
type GeneralData struct {
	Title            string
	User             string
	MessageType      string
	Message          string
	CountdownActive  bool
	CountdownSeconds int
}

// FlashMessage holds type and content of flash message displayed to the user
type FlashMessage struct {
	Type    string
	Message string
}
