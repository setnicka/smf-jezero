package game

import (
	"time"
)

type State struct {
	Teams          []team
	Rounds         []*roundState // Round i = state after i-th round (round 0 = start state)
	CurrentActions map[string]int
}

type team struct {
	Name     string
	Login    string
	Salt     string
	Password string
}

type roundState struct {
	Number      int
	GlobalState int
	Teams       map[string]teamState
	Time        time.Time
}

type teamState struct {
	Action  int
	Message string
	Money   int
}

////////////

type checkFunc func(globalState int, money int) bool
type actionFunc func(globalState int, money int, actions map[string]int) (int, int, string)

type ActionDef struct {
	DisplayName  string
	DisplayClass string
	check        checkFunc
	action       actionFunc
}
