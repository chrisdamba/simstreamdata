package models

import (
	"math/rand"
	"time"
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

func InitializeStates() *StateMachine {
	sm := NewStateMachine()
	home := &State{Page: "Home", StatusCode: 200, Method: "GET", UserLevel: "free", AuthStatus: "Logged In", EventTime: time.Now()}
	playVideo := &State{Page: "PlayVideo", StatusCode: 200, Method: "PUT", UserLevel: "free", AuthStatus: "Logged In", EventTime: time.Now()}

	sm.AddState(home)
	sm.AddState(playVideo)

	home.AddTransition(playVideo, 0.8) // 80% probability to transition from Home to PlayVideo
	playVideo.AddTransition(home, 0.5) // 50% probability to return to Home

	return sm
}
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