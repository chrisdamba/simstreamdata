package models

import (
	"math/rand"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
)

type State struct {
	Page          string
	StatusCode    int
	Method        string
	UserLevel     string
	AuthStatus    string
	Transitions   []Transition
	EventTime 	  time.Time
}

type Transition struct {
	State       *State
	Probability float64
}

type StateMachine struct {
	States       []*State
	CurrentState *State
}

func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

func (sm *StateMachine) AddState(state *State) {
	sm.States = append(sm.States, state)
}

// SetInitialState sets the initial state of the state machine.
func (sm *StateMachine) SetInitialState(state *State) {
	sm.CurrentState = state
}

// UpdateState moves the state machine to the next state based on the probability transitions.
func (sm *StateMachine) UpdateState(rng *rand.Rand) {
	if sm.CurrentState != nil {
			nextState := sm.CurrentState.GetNextState(rng)
			if nextState != nil {
					sm.CurrentState = nextState
			}
	}
}

func (s *State) AddTransition(target *State, probability float64) {
	s.Transitions = append(s.Transitions, Transition{State: target, Probability: probability})
}

func (s *State) GetNextState(rng *rand.Rand) *State {
	if len(s.Transitions) == 0 {
		return nil
	}
	p := rng.Float64()
	total := 0.0
	for _, t := range s.Transitions {
		total += t.Probability
		if p < total {
			return t.State
		}
	}
	return nil // Return nil if no transition is found (or default state)
}

func InitializeStates(cfg *config.Config) *StateMachine {
	sm := NewStateMachine()
	stateMap := make(map[string]*State) // Temporary map to store states for easy lookup

	// Step 1: Create states
	for _, page := range cfg.NewSessionPages {
		state := &State{
			Page:       page.Page,
			StatusCode: page.Status,
			Method:     page.Method,
			UserLevel:  page.Level,
			AuthStatus: page.Auth,
			EventTime:  time.Now(),
		}
		sm.AddState(state)
		stateMap[page.Page] = state // Add to map for easy reference
	}

	// Step 2: Setup transitions based on the transitions configuration
	for _, trans := range cfg.Transitions {
		sourceState := stateMap[trans.Source.Page]
		destState := stateMap[trans.Dest.Page]
		if sourceState != nil && destState != nil {
			sourceState.AddTransition(destState, trans.P)
		}
	}

	// Optionally set an initial state, here setting the first state as initial if available
	if len(sm.States) > 0 {
		sm.SetInitialState(sm.States[0])
	}

	return sm
}


// func InitializeStates(cfg *config.Config) *StateMachine {
// 	sm := NewStateMachine()
// 	home := &State{Page: "Home", StatusCode: 200, Method: "GET", UserLevel: "free", AuthStatus: "Logged In", EventTime: time.Now()}
// 	playVideo := &State{Page: "PlayVideo", StatusCode: 200, Method: "PUT", UserLevel: "free", AuthStatus: "Logged In", EventTime: time.Now()}

// 	sm.AddState(home)
// 	sm.AddState(playVideo)

// 	home.AddTransition(playVideo, 0.8) // 80% probability to transition from Home to PlayVideo
// 	playVideo.AddTransition(home, 0.5) // 50% probability to return to Home

// 	return sm
// }
/*
func initializeStates() *StateMachine {
	sm := NewStateMachine()
	// Define example states
	browsing := &State{Page: "Browse", StatusCode: 200, Method: "GET", UserLevel: "free", AuthStatus: "Logged In"}
	playback := &State{Page: "PlayVideo", StatusCode: 200, Method: "GET", UserLevel: "free", AuthStatus: "Logged In"}
	pause := &State{Page: "PauseVideo", StatusCode: 200, Method: "PUT", UserLevel: "free", AuthStatus: "Logged In"}
	accountActivity := &State{Page: "AccountUpdate", StatusCode: 200, Method: "POST", UserLevel: "free", AuthStatus: "Logged In"}

	sm.AddState(browsing)
	sm.AddState(playback)
	sm.AddState(pause)
	sm.AddState(accountActivity)

	// Define transitions
	sm.AddTransition(browsing, playback, 0.5)   // 50% chance to move from browsing to playback
	sm.AddTransition(playback, pause, 0.3)      // 30% chance to pause while playing
	sm.AddTransition(pause, playback, 0.7)      // 70% chance to resume playback
	sm.AddTransition(playback, browsing, 0.2)   // 20% chance to return to browsing after playback
	sm.AddTransition(accountActivity, browsing, 0.9) // 90% chance to go back to browsing after account activity

	return sm
}
*/