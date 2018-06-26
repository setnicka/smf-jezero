package game

import (
	"fmt"
	"math"
	"strings"
)

// Game mechanic

type checkFunc func(globalState int, money int) bool
type actionFunc func(globalState int, money int, actions map[string]int) (int, int, string)

type actionDef struct {
	displayName  string
	displayClass string
	check        checkFunc
	action       actionFunc
}

const (
	// Game constants
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
)

func inActions(action int, actions map[string]int) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////
// ACTIONS /////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

var actions = map[int]actionDef{
	ACTION_NOTHING: {
		displayName:  "Nic",
		displayClass: "",
	},

	ACTION_ECO: {
		displayName:  "Ekologická výroba",
		displayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			return -ECO_POLLUTION, globalState, fmt.Sprintf("Ekologická výroba úspěšná, získáno %d peněz a jezero znečištěno o %d jednotek", globalState, ECO_POLLUTION)
		},
	},

	ACTION_HARVEST: {
		displayName:  "Neekologická výroba",
		displayClass: "",
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
		displayName:  "Čištění",
		displayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			cleaning := CLEANING_ABSOLUTE
			if globalState > 0 {
				cleaning = cleaning - int(math.Round(float64(globalState)/float64(CLEANING_RELATIVE)))
			}

			return cleaning, 0, fmt.Sprintf("Jezero vyčištěno o %d jednotek", cleaning)
		},
	},

	ACTION_CONTROL: {
		displayName:  "Kontrola",
		displayClass: "",
		action: func(globalState int, money int, actions map[string]int) (int, int, string) {
			return 0, 0, fmt.Sprintf("Požádali jsme ministerstvo o kontrolu, pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán")
		},
	},

	ACTION_SPIONAGE: {
		displayName:  "Špionáž",
		displayClass: "",
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
