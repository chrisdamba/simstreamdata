package models

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
	"github.com/google/uuid"
)

// User holds the data for a simulated user.
type User struct {
	ID               uuid.UUID
	Alpha            float64
	Beta             float64
	StartTime        time.Time
	StateMachine    *StateMachine
	Auth             string
	InitialLevel     string
	Properties       map[string]interface{}
	Device           map[string]interface{}
	PreferredGenres  []string
	FavoriteShows    []string
	GenrePreferences map[string]int // weight for each genre
	ViewingHours     int
	SubscriptionType SubscriptionType
	CurrentSession   *Session
}

// NewUser creates a new User instance.
func NewUser(alpha, beta float64, startTime time.Time, auth, level string, subscriptionType SubscriptionType, stateMachine *StateMachine, genres map[string]int) *User {
	return &User{
		ID:               uuid.New(),
		Alpha:            alpha,
		Beta:             beta,
		StartTime:        startTime,
		Auth:             auth,
		InitialLevel:     level,
		SubscriptionType: subscriptionType,
		StateMachine: stateMachine,
		Properties:       make(map[string]interface{}),
		Device:           make(map[string]interface{}),
		GenrePreferences: genres, 
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
			u.CurrentSession.StateMachine.UpdateState(rng)
			u.CurrentSession.IncrementEvent()
	}

	// Use the Serialize method to get a consistent JSON string for logging
	eventData := u.Serialize()

	return eventData, nil
}

// startNewSession initializes a new session for the user.
func (u *User) startNewSession(rng *rand.Rand, config *config.Config) {
	// Ensure there are session pages to choose from
	if len(config.NewSessionPages) == 0 {
			panic("no session pages available in configuration")
	}


	// Assume a default engagement level, or compute it based on some logic
	engagementLevel := 0  // Placeholder for actual logic

	// Create a new session with the selected state
	u.CurrentSession = NewSession(u.ID.String(), u.StateMachine, u.SubscriptionType, engagementLevel, u.StartTime, rng, config)
}

// Serialize serializes the user's current state to a JSON string for logging.
func (u *User) Serialize() string {
	currentState := u.CurrentSession.StateMachine.CurrentState
	data := map[string]interface{}{
		"ts":              currentState.EventTime.UnixMilli(),  // UNIX milliseconds 
		"userId":          u.ID.String(),
		"sessionId":       u.CurrentSession.ID,
		"page":            currentState.Page,
		"auth":            currentState.AuthStatus,
		"method":          currentState.Method,
		"status":          currentState.StatusCode,
		"itemInSession":   u.CurrentSession.NextEventNumber, // Adapt based on your counter 
		"preferredGenres": u.PreferredGenres,
		"favoriteShows":   u.FavoriteShows,
		"viewingHours":    		u.ViewingHours,
		"subscriptionType": 	string(u.SubscriptionType),
		"deviceType": 				u.Device["type"],
		"eventType": 					u.CurrentSession.NextEventType,
		"sessionDuration": 		time.Since(u.CurrentSession.StartTime).Minutes(),
		"adsShown": 					u.CurrentSession.CurrentAd != nil,
		"videoID":     				u.CurrentSession.CurrentVideo.ID,
		"title":        			u.CurrentSession.CurrentVideo.PrimaryTitle,
		"duration":     			u.CurrentSession.CurrentVideo.RuntimeMinutes.String(),

	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error serializing user data:", err)
		return "{}"
	}
	return string(jsonData)
}

// AdjustGenrePreferences updates the user's preferences based on the genres of the recently watched video.
func (u *User) AdjustGenrePreferences(watchedVideo *config.Video) {
	for _, genre := range watchedVideo.Genres {
			if _, exists := u.GenrePreferences[genre]; exists {
					u.GenrePreferences[genre] += 1 // Increment preference weight for each watched genre
			} else {
					u.GenrePreferences[genre] = 1 // Set initial preference if genre is new to the user
			}
	}
}

// Update preferences after a video is watched
func (u *User) WatchVideo(video *config.Video) {
	u.AdjustGenrePreferences(video)
	// Additional logic for handling the watch event
}

func (u *User) DecidesToContinueWatching() bool {
	// Simulate the decision process, could use randomness or user preferences
	return rand.Float32() < 0.8 // 80% chance to continue watching
}

func (u *User) SelectVideo(config *config.Config) *config.Video {
	/*
	weightedVideos := make([]*config.Video, 0)
	weights := make([]int, 0)

	// Create a weighted list of videos based on user's genre preferences
	for _, v := range config.AllVideos {
			genreWeight := 0
			for _, genre := range v.Genres {
					if weight, ok := u.GenrePreferences[genre]; ok {
							genreWeight += weight
					}
			}
			if genreWeight > 0 { // Only consider videos that match the user's genre preferences
					weightedVideos = append(weightedVideos, &v) // Note the address of v here
					weights = append(weights, genreWeight)
			}
	}

	// Perform weighted random selection
	if len(weightedVideos) == 0 {
			return nil // No videos match the user's preferences
	}
	return weightedRandomSelect(weightedVideos, weights) // Make sure this function returns *config.Video
	*/

	// Directly reference a Video from config to see if the type is accessible
	if len(config.AllVideos) > 0 {
		// Randomly select a new video from the list
		newVideo := config.AllVideos[rand.Intn(len(config.AllVideos))]
		return &newVideo // Access the first video directly
	}
	return nil
}

// Implement weightedRandomSelect assuming it returns *config.Video
func weightedRandomSelect(videos []*config.Video, weights []int) *config.Video {
	// Your weighted selection logic here
	totalWeight := 0
	for _, weight := range weights {
			totalWeight += weight
	}

	r := rand.Intn(totalWeight)
	for i, weight := range weights {
			r -= weight
			if r <= 0 {
					return videos[i]
			}
	}

	return nil // In case something goes wrong
}


