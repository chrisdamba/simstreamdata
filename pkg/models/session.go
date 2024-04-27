package models

import (
	"fmt"
	"log"
	"math/rand"
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
    ID              string
    UserID          string
    StartTime       time.Time 
    LastEvent       time.Time 

    CurrentState    *State
    StateMachine    *StateMachine

    CurrentContent  *Content  
    CurrentAd       *Ad 
    CurrentVideo    *config.Video
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
	if s.CurrentState.Page == "PlayVideo" && s.CurrentContent != nil && s.CurrentContent.Type == VideoType {
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
            // Transition logic for ad impressions
            s.scheduleNextAdImpression()
        case "AdComplete":
            // Finish the ad sequence and resume content
            s.finishAdAndResumeContent()
	}
	s.scheduleNextEvent(s.NextEventType)
}

func (s *Session) finishAdAndResumeContent() {
    s.CurrentAd = nil // Clear the ad
    s.NextEventType = "Content"
    s.scheduleNextEvent("resume content")
}

func (s *Session) scheduleNextAdImpression() {
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

        // Handle next steps based on the session's logic
        s.decideNextSteps()
    }
}

// decideNextSteps decides what happens after a video ends.
func (s *Session) decideNextSteps() {
    if rand.Float64() < 0.5 { // 50% chance to pick a new video or end session
        videos := s.Config.AllVideos
        if len(videos) > 0 {
            // Randomly select a new video from the list
            newVideo := videos[rand.Intn(len(videos))]
            s.StartVideo(&newVideo)
        } else {
            // No videos available, end session
            s.EndSession()
        }
    } else {
        // Simulate ending the session
        s.EndSession()
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