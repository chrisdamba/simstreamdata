package models

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
)

// User holds the data for a simulated user.
type User struct {
	Alpha            float64
	Beta             float64
	StartTime        time.Time
	InitialStates    map[string]*WeightedRandomThingGenerator[*State] // Keyed by auth and level
	Auth             string
	Properties       map[string]interface{}
	Device           map[string]interface{}
	InitialLevel     string
	PreferredGenres  []string
	FavoriteShows    []string
	ViewingHours     int
	SubscriptionType SubscriptionType
	CurrentSession   *Session
}

// NewUser creates a new User instance.
func NewUser(alpha, beta float64, startTime time.Time, auth, level string, subscriptionType SubscriptionType) *User {
	return &User{
		Alpha:            alpha,
		Beta:             beta,
		StartTime:        startTime,
		Auth:             auth,
		InitialLevel:     level,
		SubscriptionType: subscriptionType,
		InitialStates:    make(map[string]*WeightedRandomThingGenerator[*State]),
		Properties:       make(map[string]interface{}),
		Device:           make(map[string]interface{}),
	}
}

// NextEvent processes the next event based on the user's current state.
func (u *User) NextEvent(rng *rand.Rand, config *config.Config) {
	if u.CurrentSession == nil || u.CurrentSession.IsDone() {
		u.startNewSession(rng)
	} else {
		u.CurrentSession.IncrementEvent(rng, config)
	}
}

// startNewSession initializes a new session for the user.
func (u *User) startNewSession(_ *rand.Rand) {
	u.CurrentSession = NewSession(u.Auth, u.SubscriptionType, u.ViewingHours, u.StartTime)
}

// Serialize serializes the user's current state to a JSON string for logging.
func (u *User) Serialize() string {
	data := map[string]interface{}{
		"userId":           u.Auth,
		"sessionPage":      u.CurrentSession.CurrentState.Page,
		"sessionAuth":      u.CurrentSession.CurrentState.AuthStatus,
		"preferredGenres":  u.PreferredGenres,
		"favoriteShows":    u.FavoriteShows,
		"viewingHours":     u.ViewingHours,
		"subscriptionType": u.SubscriptionType,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error serializing user data:", err)
		return "{}"
	}
	return string(jsonData)
}
