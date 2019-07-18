package game

import (
	"fmt"
	"html/template"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/pkg/errors"
)

// Game mechanic

const (
	tcp_visualizator = "192.168.1.101:4242"

	// Game constants
	DEFAULT_GLOBAL_STATE = 100
	DEFAULT_MONEY        = 100

	ECO_POLLUTION = 1 // znečištění po ekologické výrobě

	HARVEST_POLLUTION = 7 // znečištění po neekologické výrobě
	HARVEST_BONUS     = 100
	HARVEST_PENALTY   = 100

	CLEANING_ABSOLUTE = 10
	CLEANING_RELATIVE = 20

	SPIONAGE_COST = 25

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

// InitGame resets the whole game (deletes all rounds).
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

// GetLastState returns last round of the game.
func (s *State) GetLastState() *roundState {
	if len(s.Rounds) > 0 {
		return s.Rounds[len(s.Rounds)-1]
	} else {
		return nil
	}
}

// GetRoundNumber returns number of the current round.
func (s *State) GetRoundNumber() int {
	return s.GetLastState().Number + 1 // actual round number = last round number + 1
}

// EndRound will end this round and compute the actions.
func (s *State) EndRound() error {
	// 1. Get previous state
	previous := s.GetLastState()
	if previous == nil {
		return errors.New("No previous state, cannot calculate new round")
	}
	// 2. Calculate and save
	s.Rounds = append(s.Rounds, s.calculateRound(previous, s.CurrentActions))
	// 3. Reset actions for the next round
	for _, team := range s.Teams {
		s.CurrentActions[team.Login] = DEFAULT_ACTION
	}
	// 4. Save state
	s.Save()

	// 5. Send data over TCP
	return s.SendState()
}

// SendState sends the state to the visualizator.
func (s *State) SendState() error {
	if conn, err := net.DialTimeout("tcp", tcp_visualizator, time.Second); err == nil {
		defer conn.Close()
		fmt.Fprintf(conn, strconv.Itoa(s.GetLastState().GlobalState)+"\n")
		fmt.Fprintf(conn, "k"+strconv.Itoa(s.GetRoundNumber())+"\n")
		return nil
	} else {
		return err
	}
}

func (s *State) calculateRound(previousRound *roundState, actions map[string]int) *roundState {
	roundNumber := previousRound.Number + 1

	// 1. Prepare new round struct
	newRound := &roundState{
		Number:      roundNumber,
		GlobalState: previousRound.GlobalState,
		Teams:       map[string]teamState{},
		Time:        time.Now(),
	}

	// 2. Do actions for all teams
	for _, team := range s.Teams {
		// 2.1 Get previous state of money and current action of each team
		previousMoney := previousRound.getMoney(team.Login)
		actionID, actionDef := getAction(actions, team.Login)

		// 2.2 Execute action func
		var globalDiff, teamMoneyDiff int
		var message string
		if actionDef.action != nil {
			globalDiff, teamMoneyDiff, message = actionDef.action(s, previousRound.GlobalState, previousMoney, actions)
		}

		// 2.3 Save results
		log.Debugf("[Round %d - team '%s'] Action '%s': Global state change %d, money change %d", newRound.Number, team.Name, actionDef.DisplayName, globalDiff, teamMoneyDiff)
		newRound.GlobalState += globalDiff
		newRound.Teams[team.Login] = teamState{
			Action:  actionID,
			Money:   previousMoney + teamMoneyDiff,
			Message: template.HTML(message),
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

var actionsNames = map[int]string{
	ACTION_NOTHING:  "Nic",
	ACTION_ECO:      "Rybolov",
	ACTION_HARVEST:  "Průmyslový rybolov",
	ACTION_CLEANING: "Čištění",
	ACTION_CONTROL:  "Kontrola",
	ACTION_SPIONAGE: "Špionáž",
}

var actionsDef = map[int]ActionDef{
	ACTION_NOTHING: {
		DisplayName:  actionsNames[ACTION_NOTHING],
		DisplayClass: "",
	},

	ACTION_ECO: {
		DisplayName:  actionsNames[ACTION_ECO],
		DisplayClass: "",
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			return -ECO_POLLUTION, globalState, fmt.Sprintf("Provedli jste normální rybolov, získali jste %d ZEĎcoinů a zhoršili stav jezera o %d", globalState, ECO_POLLUTION)
		},
	},

	ACTION_HARVEST: {
		DisplayName:  actionsNames[ACTION_HARVEST],
		DisplayClass: "",
		check:        func(globalState int, money int) bool { return (money >= 0) },
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			// If there were control -> penalty
			if inActions(ACTION_CONTROL, actions) {
				return -HARVEST_POLLUTION, -HARVEST_PENALTY, fmt.Sprintf("Váš průmyslový rybolov byl odhaleno kontrolou! Nic jste nezískali a musíte místo toho zaplatit pokutu %d ZEĎcoinů", HARVEST_PENALTY)
			} else {
				gatheredMoney := globalState + HARVEST_BONUS
				return -HARVEST_POLLUTION, gatheredMoney, fmt.Sprintf("Provedli jste průmyslový rybolov, získali jste za to %d ZEĎcoinů a zhoršili stav jezera o %d", gatheredMoney, HARVEST_POLLUTION)
			}
		},
	},

	ACTION_CLEANING: {
		DisplayName:  actionsNames[ACTION_CLEANING],
		DisplayClass: "",
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			cleaning := CLEANING_ABSOLUTE
			if globalState > 0 {
				cleaning = cleaning - int(math.Round(float64(globalState)/float64(CLEANING_RELATIVE)))
			}
			return cleaning, 0, fmt.Sprintf("Zlepšili jste čištěním stav jezera o %d", cleaning)
		},
	},

	ACTION_CONTROL: {
		DisplayName:  actionsNames[ACTION_CONTROL],
		DisplayClass: "",
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			return 0, 0, fmt.Sprintf("Požádali jsme ministerstvo o kontrolu, pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán")
		},
	},

	ACTION_SPIONAGE: {
		DisplayName:  actionsNames[ACTION_SPIONAGE],
		DisplayClass: "",
		check:        func(globalState int, money int) bool { return (money >= 0) },
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			// If there were control -> no spionage
			if inActions(ACTION_CONTROL, actions) {
				return 0, -SPIONAGE_COST, fmt.Sprintf("Špionáž nemohla být dokončena kvůli probíhající kontrole jiného týmu, nic jste nezjistili")
			} else {
				results := []string{}
				for team, action := range actions {
					if s.GetTeam(team) != nil {
						results = append(results, fmt.Sprintf("<li>%s: <b>%s</b></li>", s.GetTeam(team).Name, actionsNames[action]))
					}
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

// Check if the action could be performed for given global state and team amount of money.
func (a ActionDef) Check(globalState int, money int) bool {
	if a.check == nil {
		return true
	} else {
		return a.check(globalState, money)
	}
}

func (round *roundState) getMoney(login string) int {
	if team, found := round.Teams[login]; found {
		return team.Money
	}
	log.Infof("New team '%s', initializing to the default money '%d'", login, DEFAULT_MONEY)
	return DEFAULT_MONEY
}

func getAction(actions map[string]int, login string) (int, ActionDef) {
	if actionID, found := actions[login]; found {
		if action, found := actionsDef[actionID]; found {
			return actionID, action
		}
		log.Infof("Team %s - There is no action with ID '%d', fallbacking to the default action '%s'.", login, actionID, actionsDef[DEFAULT_ACTION].DisplayName)
		return DEFAULT_ACTION, actionsDef[DEFAULT_ACTION]
	}
	log.Infof("No action for team '%s', fallbacking to the default action '%s'.", login, actionsDef[DEFAULT_ACTION].DisplayName)
	return DEFAULT_ACTION, actionsDef[DEFAULT_ACTION]
}
