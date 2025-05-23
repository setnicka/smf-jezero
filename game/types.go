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
	variant Variant
	actions map[ActionID]ActionDef

	Teams          []Team
	Rounds         []*RoundState // Round i = state after i-th round (round 0 = start state)
	CurrentActions map[string]ActionID
}

// GlobalState is number (or numbers) representing state of the jezero
// type GlobalState int
type GlobalState map[GamePart]int

// Hash for automatic checking
func (g GlobalState) Hash() string {
	parts := []string{}
	for _, part := range GameParts {
		parts = append(parts, fmt.Sprintf("%s%d", part, g[part]))
	}
	return strings.Join(parts, "-")
}

func (g GlobalState) Pretty(symbol string) string {
	parts := []string{}
	for _, part := range GameParts {
		parts = append(parts, fmt.Sprintf("%s:%d%s", part, g[part], symbol))
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
	Action  ActionID
	Message template.HTML
	Money   int
}

////////////

type checkFunc func(globalState int, money int) bool
type actionFunc func(s *State, globalState int, money int, actions map[string]ActionID) (int, int, string)

type ActionDef struct {
	DisplayName  string
	DisplayClass string
	check        checkFunc
	action       actionFunc
}

//////////

// Variant is interface for game variants (translations)
type Variant interface {
	ViewTemplateName() string
	TemplateStateName() string
	TemplateStateSymbol() string
	TemplateMoneyName() string
	TemplateMoneySymbol() string

	NopName() string

	EcoName() string
	EcoMessage(money int, pollution int) string

	HarvestName() string
	HarvestPenaltyMessage(penalty int) string
	HarvestSuccessMessage(money int, pollution int) string

	CleaningName() string
	CleaningMessage(cleaning int) string

	InspectionName() string
	InspectionMessage() string

	EspionageName() string
	EspionageFailMessage() string
	EspionageSuccessMessage(teamActions map[string]string) string

	GlobalMessage(reduce string, increase string, change int) string
}
