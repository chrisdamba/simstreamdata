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

type EventMessage struct {
	Topic   string
	Message []byte
}

type PageViewEvent struct {
	Timestamp      int64  `json:"ts"`
	SessionID      string `json:"sessionId"`
	SessionDuration float64 `json:"sessionDuration"`
	Page           string `json:"page"`
	Auth           string `json:"auth"`
	Method         string `json:"method"`
	Status         int    `json:"status"`
	UserID         string `json:"userId"`
	DeviceType     string `json:"deviceType"`
	DeviceOS   		 string `json:"deviceOs"`
	ItemInSession  int    `json:"itemInSession"`
	SubscriptionType string `json:"subscriptionType"`
}

type AuthEvent struct {
	PageViewEvent
	Success    bool `json:"success"`
}

type ListenEvent struct {
	PageViewEvent
	SongID     	string `json:"songId"`
	AudioTitle  string `json:"audioTitle"`
	ArtistName	string `json:"artistName"`
	Duration    int 	 `json:"duration"` // in seconds
}

type WatchEvent struct {
	PageViewEvent
	VideoID     string `json:"videoId"`
	VideoTitle  string `json:"videoTitle"`
	Genres 			string `json:"genres"`
	Duration    int    `json:"duration"`// in seconds
}

type AdEvent struct {
	PageViewEvent
	AdID        string `json:"adId"`
	AdType      string `json:"adType"`
	Duration    int 	 `json:"duration"`// in seconds
}

type StatusChangeEvent struct {
	PageViewEvent
	OldStatus   string
	NewStatus   string
}

// DeviceTypes defines possible types of devices for the simulation.
var DeviceTypes = []string{"smartphone", "tablet", "desktop", "laptop"}

// OperatingSystems defines possible operating systems for the devices.
var OperatingSystems = []string{"Android", "iOS", "Windows", "macOS", "Linux"}

// NewUser creates a new User instance.
func NewUser(alpha, beta float64, startTime time.Time, auth, level string, subscriptionType SubscriptionType, stateMachine *StateMachine, genres map[string]int) *User {
	// Randomly select device type and operating system for the user
	deviceType := DeviceTypes[rand.Intn(len(DeviceTypes))]
	operatingSystem := OperatingSystems[rand.Intn(len(OperatingSystems))]
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
		Device: map[string]interface{}{
			"type":    deviceType,
			"os":      operatingSystem,
			"version": "1.0", // Example fixed value for all devices
		},
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

// Serialize data into JSON for simplicity
func serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// NextEvent processes the next event for the user and returns the event data and any error encountered.
func (u *User) NextEvent(rng *rand.Rand, config *config.Config) (EventMessage, error) {
	if u.CurrentSession == nil || u.CurrentSession.IsDone() {
		u.startNewSession(rng, config)
	} else {
		u.CurrentSession.StateMachine.UpdateState(rng)
		u.CurrentSession.IncrementEvent()
		if u.CurrentSession.NextEventType == "Content" && u.CurrentSession.CurrentAd == nil {
			// Decide on post-roll ads or end session
			if u.CurrentSession.ShouldContinueSession() {
				u.CurrentSession.HandleNextVideoEvent(config)
			} else {
				u.CurrentSession.EndSession()
			}
		}
	}

	// Use the Serialize method to get a consistent JSON string for logging
	return u.Serialize(rng)
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
func (u *User) Serialize(rng *rand.Rand) (EventMessage, error) {
	currentState := u.CurrentSession.StateMachine.CurrentState  

	baseEvent := PageViewEvent{
		Timestamp:      time.Now().Unix(),
		SessionID:      u.CurrentSession.ID,
		SessionDuration: time.Since(u.CurrentSession.StartTime).Minutes(),
		Page:           currentState.Page,
		Auth:           currentState.AuthStatus,
		Method:         currentState.Method,
		Status:         currentState.StatusCode,
		UserID:         u.ID.String(),
		DeviceType:     u.Device["type"].(string),
		DeviceOS:       u.Device["os"].(string),
		ItemInSession:  u.CurrentSession.NextEventNumber,
		SubscriptionType: string(u.SubscriptionType),
	}
	/*
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
	*/

	var topic = "page_views_events"
	var event interface{}
	switch currentState.Page {
		case "Login", "Logout", "Register":
			event = AuthEvent{
				PageViewEvent: baseEvent,
				Success:       currentState.AuthStatus == "Logged In",
			}
			topic = "auth_events"

		case "PlayVideo":
			event = WatchEvent{
				PageViewEvent: baseEvent,
				VideoID:    u.CurrentSession.CurrentVideo.ID,
				VideoTitle: u.CurrentSession.CurrentVideo.PrimaryTitle,
				Duration:   int(u.CurrentSession.CurrentVideo.RuntimeMinutes),
			}
			topic = "watch_events"

		case "NextSong":
			event = ListenEvent{
				PageViewEvent: baseEvent,
				SongID:        "someSongID",
				AudioTitle:    "Some Song Title",
				ArtistName:    "Some Artist",
				Duration:      180, // example duration in seconds
			}
			topic = "listen_events"
		case "AdStart", "AdImpression", "AdEnd":
			event = AdEvent{
				PageViewEvent: baseEvent,
				AdID:       u.CurrentSession.CurrentAd.ID,
				AdType:     u.CurrentSession.CurrentAd.Type,
				Duration:   int(u.CurrentSession.CurrentAd.Duration),
			}
			topic = "ad_events"

		case "SubmitUpgrade", "SubmitDowngrade", "CancelSubscription":
			event = StatusChangeEvent{
				PageViewEvent: baseEvent,
				OldStatus:  string(u.SubscriptionType),
				NewStatus:  string(u.SubscriptionType),
			}
		default:
			event = baseEvent
	}
	// Serialize the event
	data, err := serialize(event)
	if err != nil {
		return EventMessage{}, fmt.Errorf("error serializing event: %w", err)
	}

	return EventMessage{Topic: topic, Message: data}, nil
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


