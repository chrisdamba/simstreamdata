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
	States       		[]*State
	CurrentState 		*State
	StateGenerator 	*WeightedRandomThingGenerator[*State]
}

type AuthLevelStateMap struct {
	Generators map[string]*WeightedRandomThingGenerator[*State]
}


func NewStateMachine(rng *rand.Rand) *StateMachine {
	return &StateMachine{
			StateGenerator: NewWeightedRandomThingGenerator[*State](),
	}
}

func (sm *StateMachine) AddState(state *State, weight int) {
	sm.States = append(sm.States, state)
	sm.StateGenerator.Add(state, weight)
}

// SetInitialState sets the initial state of the state machine.
func (sm *StateMachine) SetInitialState(state *State) {
	sm.CurrentState = state
}

func (sm *StateMachine) SetRandomInitialState(rng *rand.Rand, cfg *config.Config, stateMap map[string]*State) {
	if sm.StateGenerator != nil {
		initialState := sm.StateGenerator.RandomThing(rng)
		// Check if initial state has transitions
		if len(initialState.Transitions) == 0 {
			// Find all transitions where the initial state is the source
			for _, trans := range cfg.Transitions {
				if trans.Source.Page == initialState.Page {
						destState := stateMap[trans.Dest.Page]
						if destState != nil {
							initialState.AddTransition(destState, trans.P)
						}
				}
			}
		}
		sm.CurrentState = initialState
	}
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

func NewAuthLevelStateMap() *AuthLevelStateMap {
	return &AuthLevelStateMap{
		Generators: make(map[string]*WeightedRandomThingGenerator[*State]),
	}
}

func (alm *AuthLevelStateMap) Add(auth, level string, state *State, weight int) {
	key := auth + "|" + level // Simplify key management
	if _, exists := alm.Generators[key]; !exists {
		alm.Generators[key] = NewWeightedRandomThingGenerator[*State]()
	}
	alm.Generators[key].Add(state, weight)
}

func (alm *AuthLevelStateMap) GetRandomState(auth, level string, rng *rand.Rand) *State {
	key := auth + "|" + level
	if gen, exists := alm.Generators[key]; exists {
		return gen.RandomThing(rng)
	}
	return nil // or default state if required
}

func InitializeStatesWithAuthLevel(cfg *config.Config, rng *rand.Rand) *AuthLevelStateMap {
	stateMap := NewAuthLevelStateMap()
	tempStateMap := make(map[string]*State) 
	sm := NewStateMachine(rng)
	for _, page := range cfg.NewSessionPages {
		state := &State{
			Page:       page.Page,
			StatusCode: page.Status,
			Method:     page.Method,
			UserLevel:  page.Level,
			AuthStatus: page.Auth,
			EventTime:  time.Now(),
		}
		sm.AddState(state, page.Weight)
		stateMap.Add(page.Auth, page.Level, state, page.Weight)
		tempStateMap[page.Page] = state
	}

	for _, trans := range cfg.Transitions {
		sourceState := tempStateMap[trans.Source.Page]
		destState := tempStateMap[trans.Dest.Page]
		if sourceState != nil && destState != nil {
			sourceState.AddTransition(destState, trans.P)
		}
	}

	return stateMap
}
