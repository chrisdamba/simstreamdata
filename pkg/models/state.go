package models

import (
	"fmt"
	"log"
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
	Laterals      map[*State]float64
	Upgrades      map[*State]float64
	Downgrades    map[*State]float64
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

func NewState(page string, statusCode int, method string, userLevel string, authStatus string, eventTime time.Time) *State {
	return &State{
		Page:       page,
		StatusCode: statusCode,
		Method:     method,
		UserLevel:  userLevel,
		AuthStatus: authStatus,
		EventTime:  eventTime,
		Laterals:   make(map[*State]float64),
		Upgrades:   make(map[*State]float64),
		Downgrades: make(map[*State]float64),
	}
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

func (s *State) addTransition(target *State, probability float64, transitionMap map[*State]float64) error {
	totalProbability := 0.0
	for _, prob := range transitionMap {
		totalProbability += prob
	}
	if totalProbability+probability > 1.0 {
		return fmt.Errorf("total transition probability would exceed 100%%")
	}
	transitionMap[target] = probability
	return nil
}

func (s *State) AddLateralTransition(target *State, probability float64) error {
	return s.addTransition(target, probability, s.Laterals)
}

func (s *State) AddUpgradeTransition(target *State, probability float64) error {
	return s.addTransition(target, probability, s.Upgrades)
}

func (s *State) AddDowngradeTransition(target *State, probability float64) error {
	return s.addTransition(target, probability, s.Downgrades)
}

func (s *State) GetNextState(rng *rand.Rand) *State {
	combinedTransitions := make(map[*State]float64)
	for state, prob := range s.Laterals {
		combinedTransitions[state] += prob
	}
	for state, prob := range s.Upgrades {
		combinedTransitions[state] += prob
	}
	for state, prob := range s.Downgrades {
		combinedTransitions[state] += prob
	}
	p := rng.Float64()
	total := 0.0
	for state, prob := range combinedTransitions {
		total += prob
		if p < total {
			return state
		}
	}
	return nil // Return nil if no transition is found
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
		state := NewState(
			page.Page,
			page.Status,
			page.Method,
			page.Level,
			page.Auth,
			time.Now(),
		)
		sm.AddState(state, page.Weight)
		stateMap.Add(page.Auth, page.Level, state, page.Weight)
		tempStateMap[page.Page] = state
	}

	for _, trans := range cfg.Transitions {
		sourceState := tempStateMap[trans.Source.Page]
		destState := tempStateMap[trans.Dest.Page]
		if sourceState != nil && destState != nil {
			if err := sourceState.AddLateralTransition(destState, trans.P); err != nil {
				log.Printf("Error adding transition: %v", err)
			}			
		}
	}

	return stateMap
}
