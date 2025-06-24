package game

import (
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net"
	"time"

	"github.com/pkg/errors"
)

// Game mechanic

const (
	// Game constants
	defaultGlobalState = 100
	defaultMoney       = 100

	ecoPollution = 1 // znečištění po ekologické výrobě

	harvestPollution = 5 // znečištění po neekologické výrobě
	harvestBonus     = 100
	harvestPenalty   = 100

	cleaningAbsolute = 10
	cleaningRelative = 20

	espionageCost = 25
)

// ActionID represents one of the actions defined in game/mechanic.go
type ActionID int

const (
	// Actions constants
	actionNothing ActionID = iota
	actionEco
	actionHarvest
	actionCleaning
	actionControl
	actionEspionage

	defaultAction = actionNothing
)

////////////////////////////////////////////////////////////////////////////////
// CALCULATE ///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// InitGame resets the whole game (deletes all rounds).
func (s *State) InitGame() {
	slog.Debug("Init state of the game - adding initial round and resetting current actions.")

	// 1. Prepare init round
	initRound := &RoundState{
		Number: 0,
		GlobalState: map[PartID]int{
			PartA: defaultGlobalState,
			PartB: defaultGlobalState,
		},
		Time:  time.Now(),
		Teams: map[TeamID]teamState{},
	}

	// 2. Reset actions
	if s.CurrentActions == nil {
		s.CurrentActions = map[TeamID]ActionID{}
	}
	for _, team := range s.Teams {
		s.CurrentActions[team.ID] = defaultAction
		initRound.Teams[team.ID] = teamState{
			Money: defaultMoney,
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
	}
	return nil
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
		s.CurrentActions[team.ID] = defaultAction
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

func (s *State) calculateRound(previousRound *RoundState, actions map[TeamID]ActionID) *RoundState {
	roundNumber := previousRound.Number + 1

	// 1. Prepare new round struct
	newRound := &RoundState{
		Number:      roundNumber,
		GlobalState: previousRound.GlobalState.copy(),
		Teams:       map[TeamID]teamState{},
		Time:        time.Now(),
	}

	actionsByPart := map[PartID]map[TeamID]ActionID{
		PartA: {},
		PartB: {},
	}
	for _, team := range s.Teams {
		if action, found := actions[team.ID]; found {
			actionsByPart[team.Part][team.ID] = action
		}
	}

	// 2. Do actions for all teams
	for _, team := range s.Teams {
		// 2.1 Get previous state of money and current action of each team
		previousMoney := previousRound.getMoney(team.ID)
		actionID, actionDef := s.getAction(actions, team.ID)

		// 2.2 Execute action func
		var globalDiff, teamMoneyDiff int
		var message string
		if actionDef.action != nil {
			globalDiff, teamMoneyDiff, message = actionDef.action(s, previousRound.GlobalState[team.Part], previousMoney, actionsByPart[team.Part])
		}

		// 2.3 Save results
		slog.Info("round calculated", "round", newRound.Number,
			"team", team.ID, "action", actionDef.DisplayName,
			"global_diff", globalDiff, "money_diff", teamMoneyDiff)
		newRound.GlobalState[team.Part] += globalDiff
		newRound.Teams[team.ID] = teamState{
			Action:  actionID,
			Money:   previousMoney + teamMoneyDiff,
			Message: template.HTML(message),
		}
	}

	// 3. Compare parts
	d := math.Abs(float64(newRound.GlobalState[PartA] - newRound.GlobalState[PartB]))
	change := int(math.Round(math.Sqrt(d*d/150 + d/5)))
	newRound.PollutionAmount = change
	if change > 0 && newRound.GlobalState[PartA] > newRound.GlobalState[PartB] {
		newRound.GlobalState[PartA] -= change
		newRound.GlobalState[PartB] += change
		newRound.GlobalMessage = template.HTML(s.variant.GlobalMessage("A", "B", change))
		newRound.PollutionTo = PartA
	} else if change > 0 {
		newRound.GlobalState[PartA] += change
		newRound.GlobalState[PartB] -= change
		newRound.GlobalMessage = template.HTML(s.variant.GlobalMessage("B", "A", change))
		newRound.PollutionTo = PartB
	}

	return newRound
}

////////////////////////////////////////////////////////////////////////////////
// ACTIONS /////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// GetActions returns map of all possible actions for this Game
func (s *State) GetActions() map[ActionID]ActionDef {
	if s.actions != nil {
		return s.actions
	}

	s.actions = map[ActionID]ActionDef{
		actionNothing: {
			DisplayName:  s.variant.NopName(),
			DisplayClass: "",
		},

		actionEco: {
			DisplayName:  s.variant.EcoName(),
			DisplayClass: "",
			action: func(s *State, globalState int, _ int, _ map[TeamID]ActionID) (int, int, string) {
				return -ecoPollution, globalState, s.variant.EcoMessage(globalState, ecoPollution)
			},
		},

		actionHarvest: {
			DisplayName:  s.variant.HarvestName(),
			DisplayClass: "",
			check:        func(_ int, money int) bool { return (money >= 0) },
			action: func(s *State, globalState int, _ int, actions map[TeamID]ActionID) (int, int, string) {
				// If there were inspection -> penalty
				if inActions(actionControl, actions) {
					return -harvestPollution, -harvestPenalty, s.variant.HarvestPenaltyMessage(harvestPenalty)
				}

				gatheredMoney := globalState + harvestBonus
				return -harvestPollution, gatheredMoney, s.variant.HarvestSuccessMessage(gatheredMoney, harvestPollution)
			},
		},

		actionCleaning: {
			DisplayName:  s.variant.CleaningName(),
			DisplayClass: "",
			action: func(s *State, globalState int, _ int, _ map[TeamID]ActionID) (int, int, string) {
				cleaning := cleaningAbsolute
				if globalState > 0 {
					cleaning = cleaning - int(math.Round(float64(globalState)/float64(cleaningRelative)))
				}

				return cleaning, 0, s.variant.CleaningMessage(cleaning)
			},
		},

		actionControl: {
			DisplayName:  s.variant.InspectionName(),
			DisplayClass: "",
			action: func(s *State, _ int, _ int, _ map[TeamID]ActionID) (int, int, string) {
				return 0, 0, s.variant.InspectionMessage()
			},
		},

		actionEspionage: {
			DisplayName:  s.variant.EspionageName(),
			DisplayClass: "",
			check:        func(_ int, money int) bool { return (money >= 0) },
			action: func(s *State, _ int, _ int, actions map[TeamID]ActionID) (int, int, string) {
				// If there were control -> no espionage
				if inActions(actionControl, actions) {
					return 0, -espionageCost, s.variant.EspionageFailMessage()
				}

				teamActions := map[string]string{}
				for teamID, action := range actions {
					if team := s.GetTeam(teamID); team != nil {
						teamActions[team.Name] = s.actions[action].DisplayName
					}
				}
				return 0, -espionageCost, s.variant.EspionageSuccessMessage(teamActions)
			},
		},
	}

	return s.actions
}

////////////////////////////////////////////////////////////////////////////////

func inActions(action ActionID, actions map[TeamID]ActionID) bool {
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
	}
	return a.check(globalState, money)
}

func (round *RoundState) getMoney(teamID TeamID) int {
	if team, found := round.Teams[teamID]; found {
		return team.Money
	}
	slog.Info("New team created, settings default money.", "team", teamID, "money", defaultMoney)
	return defaultMoney
}

func (s *State) getAction(actions map[TeamID]ActionID, teamID TeamID) (ActionID, ActionDef) {
	slog := slog.With("team", teamID)

	if actionID, found := actions[teamID]; found {
		if action, found := s.GetActions()[actionID]; found {
			return actionID, action
		}
		slog.Warn("unknown action set, fallback to the default action",
			"action_id", actionID,
			"fallback_action", s.actions[defaultAction].DisplayName)
		return defaultAction, s.actions[defaultAction]
	}
	slog.Info("no action set, fallback to the default action ",
		"fallback_action", s.actions[defaultAction].DisplayName)
	return defaultAction, s.actions[defaultAction]
}
