package game

import (
	"fmt"
	"html/template"
	"math"
	"net"
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

	ECO_POLLUTION = 1 // znečištění po ekologické výrobě

	HARVEST_POLLUTION = 5 // znečištění po neekologické výrobě
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
	initRound := &RoundState{
		Number: 0,
		GlobalState: map[GamePart]int{
			PartA: DEFAULT_GLOBAL_STATE,
			PartB: DEFAULT_GLOBAL_STATE,
		},
		Time:  time.Now(),
		Teams: map[string]teamState{},
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
	s.Rounds = []*RoundState{initRound}

	// 3. Save
	s.Save()
}

// GetLastState returns last round of the game.
func (s *State) GetLastState() *RoundState {
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
	for _, notifyTarget := range s.cfg.TCPNotify {
		if conn, err := net.DialTimeout("tcp", notifyTarget, time.Second); err == nil {
			defer conn.Close()
			gs := s.GetLastState().GlobalState
			fmt.Fprintf(conn, "%d;%d\n", gs[PartA], gs[PartB])
			fmt.Fprintf(conn, "k%d\n", s.GetRoundNumber())
		} else {
			return err
		}
	}
	return nil
}

func (s *State) calculateRound(previousRound *RoundState, actions map[string]int) *RoundState {
	roundNumber := previousRound.Number + 1

	// 1. Prepare new round struct
	newRound := &RoundState{
		Number:      roundNumber,
		GlobalState: previousRound.GlobalState.copy(),
		Teams:       map[string]teamState{},
		Time:        time.Now(),
	}

	actionsByPart := map[GamePart]map[string]int{
		PartA: {},
		PartB: {},
	}
	for _, team := range s.Teams {
		if action, found := actions[team.Login]; found {
			actionsByPart[team.Part][team.Login] = action
		}
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
			globalDiff, teamMoneyDiff, message = actionDef.action(s, previousRound.GlobalState[team.Part], previousMoney, actionsByPart[team.Part])
		}

		// 2.3 Save results
		log.Debugf("[Round %d - team '%s'] Action '%s': Global state change %d, money change %d", newRound.Number, team.Name, actionDef.DisplayName, globalDiff, teamMoneyDiff)
		newRound.GlobalState[team.Part] += globalDiff
		newRound.Teams[team.Login] = teamState{
			Action:  actionID,
			Money:   previousMoney + teamMoneyDiff,
			Message: template.HTML(message),
		}
	}

	// 3. Compare parts
	d := math.Abs(float64(newRound.GlobalState[PartA] - newRound.GlobalState[PartB]))
	change := int(math.Round(math.Sqrt(d*d/150 + d/5)))
	if change > 0 && newRound.GlobalState[PartA] > newRound.GlobalState[PartB] {
		newRound.GlobalState[PartA] -= change
		newRound.GlobalState[PartB] += change
		newRound.GlobalMessage = template.HTML(fmt.Sprintf("<b>Znečištění přes úžinu:</b> Z moře B do moře A se přelilo %d znečištění.", change))
	} else if change > 0 {
		newRound.GlobalState[PartA] += change
		newRound.GlobalState[PartB] -= change
		newRound.GlobalMessage = template.HTML(fmt.Sprintf("<b>Znečištění přes úžinu:</b> Z moře A do moře B se přelilo %d znečištění.", change))
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
	ACTION_ECO:      "Tradiční pěstování okurek",
	ACTION_HARVEST:  "Průmyslové pěstování okurek",
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
			return -ECO_POLLUTION, globalState, fmt.Sprintf("Věnovali jste se tradičnímu pěstování, získáváte %d 🥒 a zhoršili jste stav moře o %d", globalState, ECO_POLLUTION)
		},
	},

	ACTION_HARVEST: {
		DisplayName:  actionsNames[ACTION_HARVEST],
		DisplayClass: "",
		check:        func(globalState int, money int) bool { return (money >= 0) },
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			// If there were control -> penalty
			if inActions(ACTION_CONTROL, actions) {
				return -HARVEST_POLLUTION, -HARVEST_PENALTY, fmt.Sprintf("Vaše průmyslové pěstování bylo odhaleno kontrolou! Nic jste nezískali a musíte místo toho zaplatit pokutu %d 🥒", HARVEST_PENALTY)
			} else {
				gatheredMoney := globalState + HARVEST_BONUS
				return -HARVEST_POLLUTION, gatheredMoney, fmt.Sprintf("Věnovali jste se průmyslovému pěstování, získali jste za to %d 🥒 a zhoršili stav moře o %d", gatheredMoney, HARVEST_POLLUTION)
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
			return cleaning, 0, fmt.Sprintf("Zlepšili jste čištěním stav moře o %d", cleaning)
		},
	},

	ACTION_CONTROL: {
		DisplayName:  actionsNames[ACTION_CONTROL],
		DisplayClass: "",
		action: func(s *State, globalState int, money int, actions map[string]int) (int, int, string) {
			return 0, 0, fmt.Sprintf("Požádali jste o kontrolu, pokud někdo v minulém kole prováděl něco špatného, tak byl potrestán")
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

func (round *RoundState) getMoney(login string) int {
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
