package models

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
)

type ContentType string 

const (
    Audio   ContentType = "audio"
    VideoType   ContentType = "video"
)

type SubscriptionType string

const (
    Free    SubscriptionType = "free"
    Basic   SubscriptionType = "basic"
    Premium SubscriptionType = "premium"
)

type Ad struct {
    ID        string
    Type      string 
    Duration  time.Duration
    StartTime time.Time
}

type Content struct {
    ID          string
    Type        ContentType
    Duration    time.Duration 
    Breakpoints []time.Duration // Mid-roll ads breakpoints for video
    StartTime   time.Time
}

type Session struct {
    ID              int64
    StartTime       time.Time 
    LastEvent       time.Time 

    Alpha           float64 // expected request inter-arrival time  
    Beta            float64 // expected session inter-arrival time
    Auth            string // Auth
    Level           string // Level
    ItemInSession   int     // ItemInSession

    CurrentState    *State
    PreviousState   *State
    StateMachine    *StateMachine
    StateMap        *AuthLevelStateMap

    CurrentContent  *Content  
    CurrentAd       *Ad 
    CurrentVideo    *config.Video
    CurrentMovie    *config.Movie
    CurrentMovieEnd time.Time
    VideoEndTime    time.Time
    
    // State tracking
    NextEventTime   time.Time
    NextEventType   string  // "Content", "AdStart", "AdImpression", "AdComplete" 
    NextEventNumber int // Add NextEventNumber field to track the number of events in the session
    LastAdTime      time.Time

    // User-related (for ad logic)
    SubscriptionTier SubscriptionType
    EngagementLevel  int
    Finished         bool
    Rng             *rand.Rand
	Config          *config.Config
}

// SessionIDCounter holds the current count for session IDs.
var sessionIDCounter int64
var lock sync.Mutex

// NextSessionID increments the session ID counter and returns the next ID.
func NextSessionID() int64 {
    lock.Lock()
    defer lock.Unlock()
    sessionIDCounter++
    return sessionIDCounter
}

func NewSession(nextEventTime time.Time, alpha float64, beta float64, stateMap *AuthLevelStateMap, auth string, level string, rng *rand.Rand, cfg *config.Config) *Session {
    var currentMovie *config.Movie
    var currentMovieEnd time.Time
    currentState := stateMap.GetRandomState(auth, level, rng)
    if currentState.Page == "NextVideo" {
        currentMovie = cfg.NextMovie() 
        currentMovieEnd = nextEventTime.Add(currentMovie.RuntimeMinutes)
    } else {
        currentMovie = nil 
        currentMovieEnd = time.Time{}
    }


    return &Session{
        ID: NextSessionID(),
        Alpha: alpha,
        Beta: beta,
        Auth: auth,
        Level: level,
        StateMap:     stateMap,
        CurrentState: currentState,
        NextEventTime: nextEventTime,
		Rng: rng,
		Config: cfg,
        Finished: false,
        CurrentMovie: currentMovie,
        CurrentMovieEnd: currentMovieEnd,
        ItemInSession: 0,
    }
}


func (s *Session) NextSession() *Session {
    nextEventTime := s.PickNextSessionStartTime(s.NextEventTime, s.Beta)

    nextSession := NewSession(nextEventTime, s.Alpha, s.Beta, s.StateMap, s.Auth, s.Level, s.Rng, s.Config)
    return nextSession
}

func (s *Session) IncrementEvent() {
    nextState := s.CurrentState.GetNextState(s.Rng)
    switch {
        case nextState == nil:
            fmt.Println("Next state is nil, marking session as finished.")
            s.Finished = true
        case nextState.StatusCode >= 300 && nextState.StatusCode <= 399:
            fmt.Println("Status code is within the range [300, 399].")
            s.NextEventTime = s.NextEventTime.Add(time.Second)
            s.CurrentState = nextState
            s.ItemInSession += 1
        case nextState.Page == "NextVideo":
            fmt.Println("Transitioning to NextVideo state.")
            if s.CurrentMovie == nil {
                fmt.Println("Starting a new movie.")
                seconds := exponentialRandomValue(s.Rng, s.Alpha)
                s.NextEventTime = s.NextEventTime.Add(time.Duration(seconds))
            } else if s.NextEventTime.Before(s.CurrentMovieEnd) {
                fmt.Println("Current movie has not ended yet.")
                s.NextEventTime = s.CurrentMovieEnd
                s.CurrentMovie = s.Config.NextMovie()
            } else {
                fmt.Println("Current movie has ended. Starting a new movie.")
                seconds := exponentialRandomValue(s.Rng, s.Alpha)
                s.NextEventTime = s.NextEventTime.Add(time.Duration(seconds))
                s.CurrentMovie = s.Config.NextMovie()
            }
            s.CurrentMovieEnd = s.NextEventTime.Add(s.CurrentMovie.RuntimeMinutes)
            s.PreviousState = s.CurrentState
            s.CurrentState = nextState
            s.ItemInSession += 1
        case nextState.Page == "AdStart":
            fmt.Println("Starting an advertisement.")
            s.startAd()
    
        case nextState.Page == "AdImpression":
            fmt.Println("Recording an ad impression.")
            s.scheduleNextAdImpression()
    
        case nextState.Page == "AdEnd":
            fmt.Println("Ad has completed.")
            s.finishAdAndResumeContent()
        default:
            fmt.Println("Default case.")
            seconds := exponentialRandomValue(s.Rng, s.Alpha)
            s.NextEventTime = s.NextEventTime.Add(time.Duration(seconds))
            s.CurrentState = nextState
            s.ItemInSession += 1
	}
}

// exponentialRandomValue returns a random value drawn from an exponential distribution with mean mu.
// This version uses a local RNG for better reproducibility and safety across different packages.
func exponentialRandomValue(rng *rand.Rand, mu float64) float64 {
	// rng.Float64() returns a float64 in [0.0,1.0)
	// Use -mu * log(1 - X) to transform a uniform random number into an exponential distribution
	return -mu * math.Log(1-rng.Float64())
}

func (s *Session) startAd() {
    adDuration := 30 * time.Second // Example ad duration
    adID := fmt.Sprintf("Ad-%d", s.Rng.Int())
    s.CurrentAd = &Ad{
        ID:        adID,
        Type:      "Standard",
        Duration:  adDuration,
        StartTime: time.Now(),
    }
    s.NextEventType = "AdImpression"
    s.NextEventTime = time.Now().Add(adDuration)
    s.LastAdTime = time.Now()    // Update the last ad time

	// Log the ad start for debugging.
	log.Printf("Starting Standard ad at %v, ID: %s\n", s.NextEventTime, adID)
}

func (s *Session) scheduleNextAdImpression() {
    // Simulating ad impression intervals and optionally ending the ad.
    if rand.Float64() < 0.8 { // Example probability to continue ad impressions
        s.NextEventTime = time.Now().Add(5 * time.Second) // Next impression
    } else {
        s.NextEventTime = time.Now().Add(5 * time.Second) // End of ad
        s.NextEventType = "AdEnd"
    }
}

func (s *Session) finishAdAndResumeContent() {
    s.CurrentAd = nil // Clear the ad
    s.NextEventType = "NextVideo" // Resume video playback
    s.NextEventTime = time.Now().Add(1 * time.Minute) // Example delay before next content
}








func (s *Session) IncrementEvent_Old() {
    now := time.Now()

    	// Check if it's time for the next event
	if now.After(s.NextEventTime) {
		nextState := s.CurrentState.GetNextState(s.Rng)
		if nextState == nil {
			s.Finished = true // No more transitions available
			return
		}
		s.handleTransition(nextState)
	}
}

// handleTransition updates session based on the state transition
func (s *Session) handleTransition(nextState *State) {
	s.CurrentState = nextState

	switch s.CurrentState.Page {
        case "PlayVideo", "NextSong":
            s.handleContent()
        case "AdStart", "AdImpression", "AdEnd":
            s.handleAdEvent()
        default:
            nState := nextState.GetNextState(s.Rng)
            s.scheduleNextEvent(nState.Page)
	}
}

// handleContent simulates handling different content types
func (s *Session) handleContent() {
	if s.CurrentState.Page == "NextSong" && s.CurrentContent != nil && s.CurrentContent.Type == Audio {
		// Simulate time till next song
		nextDuration := time.Duration(s.Rng.ExpFloat64()*float64(s.Config.Alpha)) * time.Second
		s.scheduleNextEventAt("NextSong", nextDuration)
	}
	if s.CurrentState.Page == "PlayVideo" && s.CurrentContent != nil && s.CurrentContent.Type == VideoType {
		// Check and handle mid-roll ad insertion
		if s.shouldInsertMidRollAd(s.Config) {
			s.startAdSequence("mid-roll")
			return
		}
		nextDuration := time.Duration(s.Rng.ExpFloat64()*float64(s.Config.Alpha)) * time.Second
		s.scheduleNextEventAt("PlayVideo", nextDuration)
	}
}

// handleAdEvent manages transitions related to advertising
func (s *Session) handleAdEvent() {
	// Define ad handling logic
	switch s.NextEventType {
        case "AdStart", "":
            // Move to AdImpression
            s.CurrentAd.StartTime = time.Now()
            s.NextEventType = "AdImpression"
            s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(10)+1) * time.Second) // Ad impressions occur shortly after ad starts
        case "AdImpression":
            // Transition logic for ad impressions
            s.scheduleNextAdImpression()
        case "AdComplete":
            // Finish the ad sequence and resume content
            s.finishAdAndResumeContent()
	}
	s.scheduleNextEvent(s.NextEventType)
}


func (s *Session) scheduleNextAdImpression_Old() {
    // Move to AdComplete or next AdImpression
    if s.Rng.Float64() < 0.8 { // 80% chance to go to next impression
        s.NextEventType = "AdImpression"
        s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(10)+1) * time.Second)
    } else {
        s.NextEventType = "AdComplete"
        s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(5)+1) * time.Second)
    }
}

func (s *Session) HandleNextVideoEvent(config *config.Config) {
    currentTime := time.Since(s.CurrentContent.StartTime)

    // Check for pre-roll ad first
    if currentTime < config.AdConfig.PreRollCooldown && s.shouldInsertPreRollAd(config) {
        s.startAdSequence("pre-roll")
        return
    }

    // Check for mid-roll ads
    if s.shouldInsertMidRollAd(config) {
        s.startAdSequence("mid-roll")
        return
    }

    // If no ads are to be inserted, schedule the next content event
    s.scheduleNextEvent("PlayVideo")
}



// startAdSequence initializes an ad sequence based on the ad type (pre-roll or mid-roll).
func (s *Session) startAdSequence(adType string) {
	// Generate an ad ID and determine the ad duration based on type.
	adID := fmt.Sprintf("Ad-%d", s.Rng.Int())
	adDuration := 30 * time.Second // Simplified example: 30-second ads

	// Update the session to reflect the ad start.
	s.CurrentAd = &Ad{
		ID:        adID,
		Type:      adType,
		Duration:  adDuration,
		StartTime: time.Now(),
	}

	// Set the next event type to "AdStart" and schedule it immediately.
	s.NextEventType = "AdStart"
	s.NextEventTime = time.Now()
    s.LastAdTime = time.Now()    // Update the last ad time

	// Log the ad start for debugging.
	log.Printf("Starting %s ad at %v, ID: %s\n", adType, s.NextEventTime, adID)
}



// scheduleNextEvent schedules the next event based on the event type
func (s *Session) scheduleNextEvent(eventType string) {
	interval := time.Duration(s.Rng.Intn(5)+1) * time.Minute
	s.NextEventTime = time.Now().Add(interval)
	s.NextEventType = eventType
}

// scheduleNextEventAt schedules the next event at a specific time interval
func (s *Session) scheduleNextEventAt(eventType string, duration time.Duration) {
	s.NextEventTime = time.Now().Add(duration)
	s.NextEventType = eventType
}

// shouldInsertPreRollAd checks if a pre-roll ad should be inserted
func (s *Session) shouldInsertPreRollAd(config *config.Config) bool {
    if time.Since(s.LastAdTime) >= config.AdConfig.PreRollCooldown && rand.Float64() < config.AdConfig.PreRollFrequency {
        s.LastAdTime = time.Now() // Update the last ad time to now
        return true
    }
    return false
}

// shouldInsertMidRollAd checks if a mid-roll ad should be inserted based on breakpoints
func (s *Session) shouldInsertMidRollAd(config *config.Config) bool {
    currentTime := time.Since(s.CurrentContent.StartTime)
    for _, bp := range s.CurrentContent.Breakpoints {
        if currentTime > bp && currentTime-bp < config.AdConfig.MidRollWindow {
            return true
        }
    }
    return false
}

func (s *Session) IsDone() bool {
    // Check if the session should be considered done
    // The session is considered done if it's explicitly marked as finished,
    // or if there's no more content and no scheduled next event.
    if s.Finished {
        return true
    }
    if s.CurrentContent == nil && s.CurrentAd == nil && time.Now().After(s.NextEventTime) {
        return true
    }
    return false
}

func (s *Session) MarkAsFinished() {
    s.Finished = true
}

// StartVideo initiates the playback of a video within the session.
func (s *Session) StartVideo(video *config.Video) {
    s.CurrentVideo = video
    // Calculate the video end time based on the runtime minutes of the video
    s.VideoEndTime = time.Now().Add(video.RuntimeMinutes)

    // Logging video start for monitoring or debugging
    log.Printf("Video %s started in session %s, ends at %s", video.PrimaryTitle, s.ID, s.VideoEndTime.Format(time.RFC3339))
}

// CheckVideoProgress checks if the current video has finished playing.
func (s *Session) CheckVideoProgress() {
    // Check if there's a current video and the current time is past the video end time
    if s.CurrentVideo != nil && time.Now().After(s.VideoEndTime) {
        log.Printf("Video %s ended in session %s", s.CurrentVideo.PrimaryTitle, s.ID)

        // Video has ended, clear the current video
        s.CurrentVideo = nil


    }
}



func (s *Session) ShouldContinueSession() bool {
	// Check if the session is already marked as finished.
	if s.Finished {
		return false
	}

	// Check if there is a current video and if it has finished playing.
	if s.CurrentVideo != nil && !time.Now().After(s.VideoEndTime) {
		return true // Continue if the video is still playing.
	}

	// Example of engagement-based logic: Continue if engagement is above a threshold.
	// This could be expanded based on complex user engagement models.
	if s.EngagementLevel > 50 {
		return true
	}

	// Check if there's a time limit on the session duration.
	maxSessionDuration := 2 * time.Hour // Example: 2 hours max duration
	return time.Since(s.StartTime) < maxSessionDuration 

	// Add more conditions as needed, for example:
	// - Check user's activity patterns.
	// - Check if there are still videos in the queue to be watched.
	// - External interruptions or conditions to end the session.

	//return false // Default to ending the session if none of the conditions match.
}


// EndSession handles the session closure.
func (s *Session) EndSession() {
    log.Printf("Session %s ended", s.ID)
    // Clean up session resources or log session completion
}

// pickNextSessionStartTime generates a new session start time by adding a randomized interval.
func (s *Session) PickNextSessionStartTime(lastTimeStamp time.Time, beta float64) time.Time {
    interval := s.generateExponential(beta) + s.Config.SessionGap
    return lastTimeStamp.Add(time.Duration(interval) * time.Second)
}

// generateExponential generates values from an exponential distribution.
// Beta is the expected session inter-arrival time (mean interval between events).
func (s *Session) generateExponential(beta float64) float64 {
    return rand.ExpFloat64() / (1 / beta) // Lambda is the rate parameter, which is 1/beta.
}

// pickFirstTimeStamp generates an initial timestamp for the session start.
func (s *Session) PickFirstTimeStamp(start time.Time, beta float64) time.Time {
    // Pick random start point, iterate to steady state.
    candidate := start.Add(-time.Duration(2*beta) * time.Second)
    for {
        candidate = s.PickNextSessionStartTime(candidate, beta)
        if candidate.After(start.Add(-time.Duration(beta) * time.Second)) {
            break
        }
    }
    return candidate
}