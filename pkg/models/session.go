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

    // User-related (for ad logic)
    SubscriptionTier SubscriptionType
    EngagementLevel  int
    Finished         bool
}

func NewSession(userID string, subscriptionTier SubscriptionType, engagementLevel int, startTime time.Time) *Session {
    sessionID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
    return &Session{
         ID: sessionID,
         UserID: userID,
         SubscriptionTier: subscriptionTier,
         EngagementLevel: engagementLevel,
         StartTime: startTime,
         LastEvent: startTime,
    }
}

func (s *Session) IncrementEvent(rng *rand.Rand, config *config.Config) {
    now := time.Now()

    // Continue with the next event if the current time is past the next event time
    if now.After(s.NextEventTime) {
        switch s.NextEventType {
        case "Content":
            if s.CurrentContent != nil {
                s.handleContentEvent(rng, config)
            }
        case "AdStart", "AdImpression", "AdComplete":
            if s.CurrentAd != nil {
                s.handleAdEvent(rng, config)
            }
        }
    }

    // Determine the next event based on the current state
    if s.CurrentContent != nil && (now.After(s.NextEventTime) || s.NextEventType == "Content") {
        if s.CurrentContent.Type == Audio {
            // Handle next song
            s.handleNextAudioEvent(rng, config)
        } else if s.CurrentContent.Type == Video {
            // Handle video playback and check for ads
            s.handleNextVideoEvent(rng, config)
        }
    } else if s.CurrentAd != nil {
        // Handle ad-related transitions
        s.handleAdEvent(rng, config)
    }


}

func (s *Session) handleContentEvent(rng *rand.Rand, config *config.Config) {
    if s.CurrentContent.Type == Audio {
        s.handleNextAudioEvent(rng, config)
    } else if s.CurrentContent.Type == Video {
        if shouldInsertAd(s, config) {
            s.startAdSequence("video", rng, config)
        } else {
            s.scheduleNextEvent("video content", rng, config)
        }
    }
}

func (s *Session) handleNextAudioEvent(rng *rand.Rand, config *config.Config) {
    // Audio event logic
    if s.SubscriptionTier == Free {
        // Check probability for an ad after a song
        if rng.Float64() < config.AdConfig.AudioAdFrequency {
            s.startAdSequence("audio", rng, config)
        }
    }
    s.scheduleNextEvent("song", rng, config)
}

func (s *Session) handleNextVideoEvent(rng *rand.Rand, config *config.Config) {
    // Check for mid-roll ads in video
    if shouldInsertAd(s, config) {
        s.startAdSequence("video", rng, config)
    } else {
        s.scheduleNextEvent("video content", rng, config)
    }
}

func (s *Session) handleAdEvent(rng *rand.Rand, config *config.Config) {
    // Logic for transitioning from one ad-related event to the next
    switch s.NextEventType {
    case "AdStart":
        // Move to AdImpression
        s.CurrentAd.StartTime = time.Now()
        s.NextEventType = "AdImpression"
        s.NextEventTime = time.Now().Add(time.Duration(rng.Intn(10)+1) * time.Second) // Ad impressions occur shortly after ad starts
    case "AdImpression":
        // Move to AdComplete or next AdImpression
        if rng.Float64() < 0.8 { // 80% chance to go to next impression
            s.NextEventType = "AdImpression"
            s.NextEventTime = time.Now().Add(time.Duration(rng.Intn(10)+1) * time.Second)
        } else {
            s.NextEventType = "AdComplete"
            s.NextEventTime = time.Now().Add(time.Duration(rng.Intn(5)+1) * time.Second)
        }
    case "AdComplete":
        // Finish the ad sequence and resume content
        s.CurrentAd = nil
        s.NextEventType = "Content"
        s.scheduleNextEvent("resume content", rng, config)
    }
}

func (s *Session) startAdSequence(_ string, _ *rand.Rand, _ *config.Config) {
    // Start a sequence of ad-related events
    adType := "pre-roll" // Simplified example
    s.CurrentAd = &Ad{ID: "Ad123", Type: adType, Duration: 30 * time.Second}
    s.NextEventType = "AdStart"
    s.NextEventTime = time.Now().Add(10 * time.Second)  // Start ad in 10 seconds
}

func (s *Session) scheduleNextEvent(eventType string, rng *rand.Rand, _ *config.Config) {
    // Calculate the next event time based on content or ad logic
    interval := time.Duration(rng.Intn(5)+1) * time.Minute // Random interval between events
    s.NextEventTime = time.Now().Add(interval)
    s.NextEventType = eventType
}

func shouldInsertAd(s *Session, config *config.Config) bool {
    if s.SubscriptionTier != Free || s.CurrentContent.Type != Video {
        return false // No ads for non-free users or audio
    }

    // Implement ad insertion logic based on user tier and video breakpoints
    if s.SubscriptionTier == Free && s.CurrentContent.Type == Video {
        currentTime := time.Since(s.StartTime)
        for _, bp := range s.CurrentContent.Breakpoints {
            if currentTime > bp && currentTime-bp < time.Minute {
                return true // Insert ad at breakpoint
            }
        }
    }

    // Pre-roll Ads  
    if s.shouldInsertPreRollAd(config) { 
        return true    
    }

    // Mid-Roll Ads (only if pre-roll didn't occur)
    currentTime := time.Since(s.CurrentContent.StartTime) // Use Content Start Time
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

func (s *Session) shouldInsertPreRollAd(config *config.Config) bool {
    // Consider factors like time since the last ad
    if time.Since(s.LastAdTime) < config.AdConfig.PreRollCooldown {
        return false
    }
    
    // Probabilistic logic based on config.AdConfig.PreRollFrequency 
    if rand.Float64() < config.AdConfig.PreRollFrequency {
        s.LastAdTime = time.Now() // Update for cooldown 
        return true
    }
    return false  
}