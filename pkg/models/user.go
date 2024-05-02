package models

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
	"github.com/bxcodec/faker/v3"
)

// User holds the data for a simulated user.
type User struct {
	ID               int64
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
	Rng             *rand.Rand
	Config          *config.Config
}

// Queue interface defines the queue operations.
type Queue interface {
	Enqueue(item interface{})
	Dequeue() (item interface{}, ok bool)
}

// SliceQueue implements a queue using a slice to hold elements.
type SliceQueue struct {
	items []interface{}
}

func NewSliceQueue() *SliceQueue {
	return &SliceQueue{items: make([]interface{}, 0)}
}

func (q *SliceQueue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

func (q *SliceQueue) Dequeue() (interface{}, bool) {
	if len(q.items) == 0 {
			return nil, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// UserQueue is a queue specifically for Users.
type UserQueue struct {
	queue Queue
}

func NewUserQueue() *UserQueue {
	return &UserQueue{
		queue: NewSliceQueue(),
	}
}

func (uq *UserQueue) Enqueue(user *User) {
	uq.queue.Enqueue(user)
}

func (uq *UserQueue) Dequeue() (*User, bool) {
	item, ok := uq.queue.Dequeue()
	if !ok {
			return nil, false
	}
	return item.(*User), true
}

type EventMessage struct {
	Topic   string
	Message []byte
}

type PageViewEvent struct {
	Timestamp      int64  `json:"ts"`
	SessionID      int64 `json:"sessionId"`
	SessionDuration float64 `json:"sessionDuration"`
	Page           string `json:"page"`
	Auth           string `json:"auth"`
	Method         string `json:"method"`
	Status         int    `json:"status"`
	UserID         int64 `json:"userId"`
	DeviceType     string `json:"deviceType"`
	DeviceOS   		 string `json:"deviceOs"`
	ItemInSession  int    `json:"itemInSession"`
	SubscriptionType 	string `json:"subscriptionType"`
	FirstName 				string `json:"firstName"`
	LastName 					string `json:"lastName"`
	Gender 						string `json:"gender"`
	DateOfBirth				string `json:"dob"`
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


// SessionIDCounter holds the current count for user IDs.
var userIDCounter int64

// NextSessionID increments the user ID counter and returns the next ID.
func NextUserID() int64 {
    lock.Lock()
    defer lock.Unlock()
    userIDCounter++
    return userIDCounter
}

// NewUser creates a new User instance.
func NewUser(alpha float64, beta float64, startTime time.Time, auth, level string, cfg *config.Config, rng *rand.Rand, genres map[string]int) *User {
	// Randomly select device type and operating system for the user
	deviceType := DeviceTypes[rand.Intn(len(DeviceTypes))]
	operatingSystem := OperatingSystems[rand.Intn(len(OperatingSystems))]
	// Generate fake user details
	firstName := faker.FirstName()
	lastName := faker.LastName()
	gender := faker.Gender()
	dob := faker.Date()
	tempSession := &Session{
		Config: cfg,  
		Rng: rng,  
	}
	nextEventTime := tempSession.PickFirstTimeStamp(startTime, beta)
	stateMap := InitializeStatesWithAuthLevel(cfg, rng)
	session := NewSession(nextEventTime, alpha, beta, stateMap, auth, level, rng, cfg)

	return &User{
		ID:               NextUserID(),
		Alpha:            alpha,
		Beta:             beta,
		StartTime:        startTime,
		Auth:             auth,
		InitialLevel:     level,
		CurrentSession:  	session,
		Properties: map[string]interface{}{
			"firstName": firstName,
			"lastName": lastName,
			"gender": gender,
			"dob": dob,
		},
		Device: map[string]interface{}{
			"type":    deviceType,
			"os":      operatingSystem,
			"version": "1.0",
		},
		GenrePreferences: genres,
		Rng: 					 		rng,
		Config:          	cfg,
	}
}

// nextEvent with optional probability of attrition
func (u *User) NextEvent(prAttrition ...float64) {
	u.CurrentSession.IncrementEvent()
	if u.CurrentSession.IsDone() {
		var probability float64
		if len(prAttrition) > 0 {
			probability = prAttrition[0]
		}

		if u.Rng.Float64() < probability || u.CurrentSession.CurrentState.AuthStatus == "" {
			u.CurrentSession.NextEventTime = time.Time{} // Assign zero value of time.Time
			fmt.Println("Session marked as potentially churned")
		} else {
			u.CurrentSession = u.CurrentSession.NextSession()
			fmt.Println("Moved to next session")
		}
	}
}


// Serialize data into JSON for simplicity
func serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}


// Serialize serializes the user's current state to a JSON string for logging.
func (u *User) Serialize(rng *rand.Rand, config *config.Config) (EventMessage, error) {
	currentState := u.CurrentSession.CurrentState  

	baseEvent := PageViewEvent{
		Timestamp:      time.Now().Unix(),
		SessionID:      u.CurrentSession.ID,
		SessionDuration: time.Since(u.CurrentSession.StartTime).Minutes(),
		Page:           currentState.Page,
		Auth:           currentState.AuthStatus,
		Method:         currentState.Method,
		Status:         currentState.StatusCode,
		UserID:         u.ID,
		DeviceType:     u.Device["type"].(string),
		DeviceOS:       u.Device["os"].(string),
		ItemInSession:  u.CurrentSession.ItemInSession,
		SubscriptionType: string(u.SubscriptionType),
		FirstName: 			u.Properties["firstName"].(string),
		LastName: 			u.Properties["lastName"].(string),
		DateOfBirth: 		u.Properties["dob"].(string),	
	}

	var topic = "page_views_events"
	var event interface{}
	switch currentState.Page {
		case "Login", "Logout", "Register":
			event = AuthEvent{
				PageViewEvent: baseEvent,
				Success:       currentState.AuthStatus == "Logged In",
			}
			topic = "auth_events"

		case "NextVideo":
			event = WatchEvent{
				PageViewEvent: baseEvent,
				VideoID:    u.CurrentSession.CurrentMovie.MovieID,
				VideoTitle: u.CurrentSession.CurrentMovie.Name,
				Duration:   int(u.CurrentSession.CurrentMovie.RuntimeMinutes.Minutes()),
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
			if u.CurrentSession.CurrentAd == nil {
				u.CurrentSession.handleAdEvent()
			} 
			event = AdEvent{
				PageViewEvent: baseEvent,
				AdID:       u.CurrentSession.CurrentAd.ID,
				AdType:     u.CurrentSession.CurrentAd.Type,
				Duration:   int(u.CurrentSession.CurrentAd.Duration),
			}
			topic = "ad_events"

		case "Submit Upgrade", "Submit Downgrade", "Cancel Subscription":
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


