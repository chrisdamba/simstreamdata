package models

import (
	"math/rand"
)

type State struct {
    Page         string
    StatusCode   int
		Method       string
    UserLevel    string
    AuthStatus   string
    ShowDetails  bool
}

type Transition struct {
    From     *State
    To       *State
    Probability float64
}

type StateMachine struct {
	States          map[string]*State // Using a map with state names as keys
	Transitions 		map[*State][]Transition
	CurrentState    *State // Track the current active state
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		States:      make(map[string]*State),
		Transitions: make(map[*State][]Transition),
	}
}

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

func (sm *StateMachine) AddState(state *State) {
	sm.States[state.Page] = state
}

func (sm *StateMachine) AddTransition(from, to *State, probability float64) {
	sm.Transitions[from] = append(sm.Transitions[from], Transition{From: from, To: to, Probability: probability})
}

func (sm *StateMachine) GetNextState(current *State) *State {
	transitions := sm.Transitions[current]
	sum := 0.0
	for _, transition := range transitions {
			sum += transition.Probability
	}
	randVal := rand.Float64() * sum
	cumulative := 0.0
	for _, transition := range transitions {
			cumulative += transition.Probability
			if randVal < cumulative {
					return transition.To
			}
	}
	return nil // Or default state
}

// Helper for convenient transition setup
func (sm *StateMachine) ConnectStates(fromName, toName string, probability float64) {
    fromState := sm.States[fromName]
    toState := sm.States[toName]
    sm.AddTransition(fromState, toState, probability)
}
