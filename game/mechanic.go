package game

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/pkg/errors"
)

// Game mechanic

const (
	// Game constants
	DEFAULT_GLOBAL_STATE = 100
	DEFAULT_MONEY        = 100

	ECO_POLLUTION = 10 // znečištění po ekologické výrobě

	HARVEST_POLLUTION = 30 // znečištění po neekologické výrobě
	HARVEST_PENALTY   = 200
	HARVEST_BONUS     = 100

	CLEANING_ABSOLUTE = 20
	CLEANING_RELATIVE = 4

	SPIONAGE_COST = 20

	// Actions constants
	ACTION_NOTHING = iota
	ACTION_ECO
	ACTION_HARVEST
	ACTION_CLEANING
	ACTION_CONTROL
	ACTION_SPIONAGE

	DEFAULT_ACTION = ACTION_NOTHING
)

////////////////////////////////////////////////////////////////////////////////
// CALCULATE ///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// will reset all
func (s *State) InitGame() {
	log.Debugf("Init state of the game - adding initial round and resetting current actions")

	// 1. Prepare init round
	initRound := &roundState{
		Number:      0,
		GlobalState: DEFAULT_GLOBAL_STATE,
		Time:        time.Now(),
		Teams:       map[string]teamState{},
	}

	// 2. Reset actions
	if s.CurrentActions == nil {
		s.CurrentActions = map[string]int{}
	}
	for _, team := range s.Teams {
		s.CurrentActions[team.Login] = DEFAULT_ACTION
		initRound.Teams[team.Login] = teamState{
			Money: DEFAULT_MONEY,
		}
	}
	s.Rounds = []*roundState{initRound}

	// 3. Save
	s.Save()
}

func (s *State) GetCurrentState() *roundState {
	if len(s.Rounds) > 0 {
		return s.Rounds[len(s.Rounds)-1]
	} else {
		return nil
	}
}

func (s *State) CalculateRound() error {
	// 1. Get previous state
	previous := s.GetCurrentState()
	if previous == nil {
		return errors.New("No previous state, cannot calculate new round")
	}
	// 2. Calculate and save
	s.Rounds = append(s.Rounds, calculateRound(previous, s.CurrentActions))
	// 3. Reset actions for the next round
	for _, team := range s.Teams {
		s.CurrentActions[team.Login] = DEFAULT_ACTION
	}
	// 4. Save state
	s.Save()
	return nil
}

func calculateRound(previous *roundState, actions map[string]int) *roundState {
	roundNumber := previous.Number + 1

	// 1. Ensure that there are all actions and all previous states
	previousMoney := map[string]int{}
	for team, state := range previous.Teams {
		if _, found := actions[team]; !found {
			log.Errorf("[Round %d] Missing action for team '%s', fallback to the default action '%s'", roundNumber, actionsDef[DEFAULT_ACTION].DisplayName)
			actions[team] = DEFAULT_ACTION
		}
		previousMoney[team] = state.Money
	}
	for team, action := range actions {
		if _, found := actionsDef[action]; !found {
			log.Errorf("[Round %d] Unknown action '%d' for team '%s', fallback to the default action '%s'", roundNumber, action, team, actionsDef[DEFAULT_ACTION].DisplayName)
			actions[team] = DEFAULT_ACTION
		}
		if _, found := previousMoney[team]; !found {
			log.Errorf("[Round %d] Missing previous state for team '%s', initializing to the default money '%d'", roundNumber, team, DEFAULT_MONEY)
			previousMoney[team] = DEFAULT_MONEY
		}
	}

	// 2. Prepare new round struct
	newRound := &roundState{
		Number:      roundNumber,
		GlobalState: previous.GlobalState,
		Teams:       map[string]teamState{},
		Time:        time.Now(),
	}

	// 3. Do all actions
	for team, action := range actions {
		// We checked, that action exists, no need to check again
		actionDef, _ := actionsDef[action]

		// 3.1 Execute action func
		globalDiff, teamMoneyDiff, message := actionDef.action(previous.GlobalState, previousMoney[team], actions)

		// 3.2 Save results
		log.Debugf("[Round %d - team '%s'] Action '%s': Global state change %d, money change %d", newRound.Number, team, actionDef.DisplayName, globalDiff, teamMoneyDiff)
		newRound.GlobalState += globalDiff
		newRound.Teams[team] = teamState{
			Action:  action,
			Money:   previousMoney[team] + teamMoneyDiff,
			Message: message,
		}
	}

	return newRound
}

////////////////////////////////////////////////////////////////////////////////
// ACTIONS /////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func GetActions() map[int]ActionDef {
	return actionsDef
}

var actionsDef = map[int]ActionDef{
	ACTION_NOTHING: {
		DisplayName:  "Nic",
		DisplayClass: "",
	},

	ACTION_ECO: {
		DisplayName:  "Ekologická výroba",
		DisplayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			return -ECO_POLLUTION, globalState, fmt.Sprintf("Ekologická výroba úspěšná, získáno %d peněz a jezero znečištěno o %d jednotek", globalState, ECO_POLLUTION)
		},
	},

	ACTION_HARVEST: {
		DisplayName:  "Neekologická výroba",
		DisplayClass: "",
		check:        func(globalState int, money int) bool { return (money >= 0) },
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			// If there were control -> penalty
			if inActions(ACTION_CONTROL, actions) {
				return -HARVEST_POLLUTION, -HARVEST_PENALTY, fmt.Sprintf("Neekologická výroba byla odhalena kontrolou! Nic jste nezískali a musíte místo toho zaplatit pokutu %d peněz", HARVEST_PENALTY)
			} else {
				gathered_money := globalState + HARVEST_BONUS
				return -HARVEST_POLLUTION, gathered_money, fmt.Sprintf("Neekologická výroba úspěšná, získáno %d peněz a jezero znečištěno o %d jednotek", gathered_money, HARVEST_POLLUTION)
			}
		},
	},

	ACTION_CLEANING: {
		DisplayName:  "Čištění",
		DisplayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			cleaning := CLEANING_ABSOLUTE
			if globalState > 0 {
				cleaning = cleaning - int(math.Round(float64(globalState)/float64(CLEANING_RELATIVE)))
			}

			return cleaning, 0, fmt.Sprintf("Jezero vyčištěno o %d jednotek", cleaning)
		},
	},

	ACTION_CONTROL: {
		DisplayName:  "Kontrola",
		DisplayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			return 0, 0, fmt.Sprintf("Požádali jsme ministerstvo o kontrolu, pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán")
		},
	},

	ACTION_SPIONAGE: {
		DisplayName:  "Špionáž",
		DisplayClass: "",
		check:        func(globalState int, money int) bool { return (money >= 0) },
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			// If there were control -> no spionage
			if inActions(ACTION_CONTROL, actions) {
				return 0, -SPIONAGE_COST, fmt.Sprintf("Špionáž nemohla být dokončena kvůli probíhající kontrole jiného týmu, nic jste nezjistili")
			} else {
				results := []string{}
				for team, action := range actions {
					results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", team, action))
				}
				return 0, -SPIONAGE_COST, fmt.Sprintf("Špionáž úspěšná, zjištěno:<ul>\n%s\n</ul>", strings.Join(results, "\n"))
			}
		},
	},
}

////////////////////////////////////////////////////////////////////////////////

func inActions(action int, actions map[string]int) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}
