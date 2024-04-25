package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
)

type ContentType string 

const (
    Audio   ContentType = "audio"
    Video   ContentType = "video"
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
    ID              string
    UserID          string
    StartTime       time.Time 
    LastEvent       time.Time 

    CurrentState    *State
    StateMachine    *StateMachine

    CurrentContent  *Content  
    CurrentAd       *Ad 
    
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

func NewSession(userID string, stateMachine *StateMachine, subscriptionTier SubscriptionType, engagementLevel int, startTime time.Time, rng *rand.Rand, config *config.Config) *Session {
    sessionID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

    // Calculate the initial next event time.
    // Here we assume the initial delay for the next event is between 1 to 5 minutes.
    initialDelay := time.Duration(rng.Intn(5)+1) * time.Minute
    nextEventTime := startTime.Add(initialDelay)

    return &Session{
        ID: sessionID,
        UserID: userID,
        SubscriptionTier: subscriptionTier,
        EngagementLevel: engagementLevel,
        StartTime: startTime,
        LastEvent: startTime,
        CurrentState: stateMachine.CurrentState,
        StateMachine: stateMachine,
        NextEventTime: nextEventTime,
		Rng: rng,
		Config: config,
    }
}

func (s *Session) IncrementEvent() {
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
		s.scheduleNextEvent("generic")
	}
}

// handleContent simulates handling different content types
func (s *Session) handleContent() {
	if s.CurrentState.Page == "NextSong" && s.CurrentContent != nil && s.CurrentContent.Type == Audio {
		// Simulate time till next song
		nextDuration := time.Duration(s.Rng.ExpFloat64()*float64(s.Config.Alpha)) * time.Second
		s.scheduleNextEventAt("song", nextDuration)
	}
	if s.CurrentState.Page == "PlayVideo" && s.CurrentContent != nil && s.CurrentContent.Type == Video {
		// Check and handle mid-roll ad insertion
		if s.shouldInsertMidRollAd(s.Config) {
			s.startAdSequence("mid-roll")
			return
		}
		nextDuration := time.Duration(s.Rng.ExpFloat64()*float64(s.Config.Alpha)) * time.Second
		s.scheduleNextEventAt("video content", nextDuration)
	}
}

// handleAdEvent manages transitions related to advertising
func (s *Session) handleAdEvent() {
	// Define ad handling logic
	switch s.NextEventType {
    case "AdStart":
        // Move to AdImpression
        s.CurrentAd.StartTime = time.Now()
        s.NextEventType = "AdImpression"
        s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(10)+1) * time.Second) // Ad impressions occur shortly after ad starts
    case "AdImpression":
        // Move to AdComplete or next AdImpression
        if s.Rng.Float64() < 0.8 { // 80% chance to go to next impression
            s.NextEventType = "AdImpression"
            s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(10)+1) * time.Second)
        } else {
            s.NextEventType = "AdComplete"
            s.NextEventTime = time.Now().Add(time.Duration(s.Rng.Intn(5)+1) * time.Second)
        }
    case "AdComplete":
        // Finish the ad sequence and resume content
        s.CurrentAd = nil
        s.NextEventType = "Content"
        s.scheduleNextEvent("resume content")
	}
	s.scheduleNextEvent(s.NextEventType)
}


func (s *Session) handleContentEvent(rng *rand.Rand, config *config.Config) {
    if s.CurrentContent.Type == Audio {
        s.handleNextAudioEvent(rng, config)
    } else if s.CurrentContent.Type == Video {
        if shouldInsertAd(s, config) {
            s.startAdSequence("video")
        } else {
            s.scheduleNextEvent("video content")
        }
    }
}

func (s *Session) handleNextAudioEvent(rng *rand.Rand, config *config.Config) {
    // Audio event logic
    if s.SubscriptionTier == Free {
        // Check probability for an ad after a song
        if rng.Float64() < config.AdConfig.AudioAdFrequency {
            s.startAdSequence("audio")
        }
    }
    s.scheduleNextEvent("song")
}

func (s *Session) handleNextVideoEvent(rng *rand.Rand, config *config.Config) {
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
    s.scheduleNextEvent("video content")
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

	// Log the ad start for debugging.
	fmt.Printf("Starting %s ad at %v, ID: %s\n", adType, s.NextEventTime, adID)
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


// shouldInsertAd decides if an ad should be inserted at any point in the video content.
func shouldInsertAd(s *Session, config *config.Config) bool {
    // Directly return false if the session is not eligible for ads.
    if s.SubscriptionTier != Free || s.CurrentContent.Type != Video {
        return false // No ads for non-free users or for non-video content.
    }

    // Check for pre-roll ad opportunity.
    // Assuming pre-roll ads can only start at the beginning of a video session or content play.
    if s.CurrentContent.StartTime == s.StartTime && s.shouldInsertPreRollAd(config) {
        return true
    }

    // Check for mid-roll ad opportunities based on defined breakpoints.
    return s.shouldInsertMidRollAd(config)
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
