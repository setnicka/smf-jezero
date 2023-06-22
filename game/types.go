package game

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/setnicka/smf-jezero/config"
)

type GamePart string

const (
	PartA GamePart = "A"
	PartB GamePart = "B"
)

var GameParts = []GamePart{PartA, PartB}

type State struct {
	cfg     config.GameConfig

	Teams          []Team
	Rounds         []*RoundState // Round i = state after i-th round (round 0 = start state)
	CurrentActions map[string]int
}

// GlobalState is number (or numbers) representing state of the jezero
//type GlobalState int
type GlobalState map[GamePart]int

func (g GlobalState) String() string {
	parts := []string{}
	for _, part := range GameParts {
		parts = append(parts, fmt.Sprintf("%s:%d", part, g[part]))
	}
	return strings.Join(parts, ", ")
}

func (g GlobalState) copy() GlobalState {
	new := GlobalState{}
	for _, part := range GameParts {
		new[part] = g[part]
	}
	return new
}

func (g GlobalState) GetA() int { return g[PartA] }
func (g GlobalState) GetB() int { return g[PartB] }

type Team struct {
	Part     GamePart // to which part of the game team belongs
	Name     string
	Login    string
	Salt     string
	Password string
}

type RoundState struct {
	Number        int
	GlobalState   GlobalState
	GlobalMessage template.HTML
	Teams         map[string]teamState
	Time          time.Time
}

func (rs RoundState) RoundNumber() int {
	return rs.Number + 1
}

type teamState struct {
	Action  int
	Message template.HTML
	Money   int
}

////////////

type checkFunc func(globalState int, money int) bool
type actionFunc func(s *State, globalState int, money int, actions map[string]int) (int, int, string)

type ActionDef struct {
	DisplayName  string
	DisplayClass string
	check        checkFunc
	action       actionFunc
}
