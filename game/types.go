package game

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/setnicka/smf-jezero/config"
)

// PartID is identification of the game section
type PartID string

// Defined game parts (WARN: when adding another part, lot of game logic must be modified)
const (
	PartA PartID = "A"
	PartB PartID = "B"
)

var allParts = []PartID{PartA, PartB}

// State of the game
type State struct {
	cfg     config.GameConfig
	variant Variant
	actions map[ActionID]ActionDef

	Teams          []Team
	Rounds         []*RoundState // Round i = state after i-th round (round 0 = start state)
	CurrentActions map[TeamID]ActionID
}

// GlobalState is number (or numbers) representing state of the jezero
// type GlobalState int
type GlobalState map[PartID]int

// Hash for automatic checking
func (g GlobalState) Hash() string {
	parts := []string{}
	for _, part := range allParts {
		parts = append(parts, fmt.Sprintf("%s%d", part, g[part]))
	}
	return strings.Join(parts, "-")
}

// Pretty print of the global state
func (g GlobalState) Pretty(symbol string) string {
	parts := []string{}
	for _, part := range allParts {
		parts = append(parts, fmt.Sprintf("%s:%d%s", part, g[part], symbol))
	}
	return strings.Join(parts, ", ")
}

func (g GlobalState) copy() GlobalState {
	newState := GlobalState{}
	for _, part := range allParts {
		newState[part] = g[part]
	}
	return newState
}

// GetA is getter used from templates
func (g GlobalState) GetA() int { return g[PartA] }

// GetB is getter used from templates
func (g GlobalState) GetB() int { return g[PartB] }

// TeamID is login of the team
type TeamID string

// Team related config
type Team struct {
	Part     PartID // to which part of the game team belongs
	ID       TeamID
	Name     string
	Login    string
	Salt     string
	Password string
}

// RoundState holds global state and team states for given round
type RoundState struct {
	Number        int
	GlobalState   GlobalState
	GlobalMessage template.HTML
	Teams         map[TeamID]teamState
	Time          time.Time
}

// RoundNumber in human readable form (indexed from 1)
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
type actionFunc func(s *State, globalState int, money int, actions map[TeamID]ActionID) (int, int, string)

// ActionDef holds definition of game action
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
