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
// func (u *User) NextEvent(rng *rand.Rand, config *config.Config) {
// 	if u.CurrentSession == nil || u.CurrentSession.IsDone() {
// 		u.startNewSession(rng, config)
// 	} else {
// 		u.CurrentSession.IncrementEvent(rng, config)
// 	}
// }

// NextEvent processes the next event for the user and returns the event data and any error encountered.
func (u *User) NextEvent(rng *rand.Rand, config *config.Config) (string, error) {
	if u.CurrentSession == nil || u.CurrentSession.IsDone() {
			u.startNewSession(rng, config)
	} else {
			u.CurrentSession.IncrementEvent(rng, config)
	}

	// Serialize the current state of the user or the event data to JSON
	eventData, err := json.Marshal(u.CurrentSession) // Assuming you want to serialize the session data
	if err != nil {
			return "", fmt.Errorf("error serializing event data: %w", err)
	}

	return string(eventData), nil
}

// startNewSession initializes a new session for the user.
func (u *User) startNewSession(rng *rand.Rand, config *config.Config) {
	// Ensure there are session pages to choose from
	if len(config.NewSessionPages) == 0 {
			panic("no session pages available in configuration")
	}

	// Select a random session page from the configuration
	randomIndex := rng.Intn(len(config.NewSessionPages))
	page := config.NewSessionPages[randomIndex]

	// Create a new state based on the randomly selected session page
	state := &State{
			Page:       page.Page,
			StatusCode: page.Status,
			Method:     page.Method,
			UserLevel:  page.Level,
			AuthStatus: page.Auth,
	}

	// Assume a default engagement level, or compute it based on some logic
	engagementLevel := 0  // Placeholder for actual logic

	// Create a new session with the selected state
	u.CurrentSession = NewSession(u.Auth, state, u.SubscriptionType, engagementLevel, u.StartTime)
}

// Serialize serializes the user's current state to a JSON string for logging.
func (u *User) Serialize() string {
	data := map[string]interface{}{
		"ts":              u.CurrentSession.CurrentState.CurrentEventTime.UnixMilli(),  // UNIX milliseconds 
		"userId":          u.Auth,
		"sessionId":       u.CurrentSession.ID,
		"page":            u.CurrentSession.CurrentState.Page,
		"auth":            u.CurrentSession.CurrentState.AuthStatus,
		"method":          u.CurrentSession.CurrentState.Method,
		"status":          u.CurrentSession.CurrentState.StatusCode,
		"itemInSession":   u.CurrentSession.NextEventNumber, // Adapt based on your counter 
		"preferredGenres": u.PreferredGenres,
		"favoriteShows":   u.FavoriteShows,
		"viewingHours":    u.ViewingHours,
		"subscriptionType": string(u.SubscriptionType),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error serializing user data:", err)
		return "{}"
	}
	return string(jsonData)
}


