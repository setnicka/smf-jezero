package game

import (
	"time"
)

type State struct {
	Teams  []team
	Rounds []round // Round i = state after i-th round (round 0 = start state)
}

type team struct {
	Name     string
	Login    string
	Salt     string
	Password string
}

type round struct {
	TeamActions map[string]int
	GlobalState int
	TeamMoney   map[string]int
}
