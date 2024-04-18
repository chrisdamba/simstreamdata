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
	States      []*State
	Transitions map[*State][]Transition
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
			Transitions: make(map[*State][]Transition),
	}
}

func (sm *StateMachine) AddState(state *State) {
	sm.States = append(sm.States, state)
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
