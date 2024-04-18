package config

import (
	"encoding/json"
	"os"
	"time"
)

// ContentType defines the types of content and their distribution weights
type ContentType struct {
    Type   string `json:"type"`
    Weight int    `json:"weight"`
}

// AdEvent defines the structure for ad-related events and their weights
type AdEvent struct {
    Event  string `json:"event"`
    Weight int    `json:"weight"`
}

// AdConfig holds the configuration for advertisement behavior in audio and video streams
type AdConfig struct {
    AudioAdFrequency  float64   `json:"audio-ad-frequency"`
    VideoAdFrequency  float64   `json:"video-ad-frequency"`
    AdEvents          []AdEvent `json:"ad-events"` 
    PreRollFrequency  float64   `json:"pre-roll-ad-frequency"` 
    PreRollCooldown   time.Duration `json:"pre-roll-ad-cooldown"`  
    MidRollWindow     time.Duration `json:"mid-roll-ad-window"` 
}

// {
//     "ad-config": {
//       "audio-ad-frequency": 0.2, // 20% chance of audio ad
//       "video-ad-frequency": 0.4, 
//       "ad-events": [ /* ... your AdEvent definitions ... */ ],
//       "pre-roll-ad-frequency": 0.6, 
//       "pre-roll-ad-cooldown": 60 * time.Second, // 1 minute cooldown
//       "mid-roll-ad-window": 30 * time.Second  // 30 seconds for mid-roll
//     }
//   }
  
type Preference struct {
    Name   string `json:"name"`
    Weight int    `json:"weight"`
}

type SubscriptionChance struct {
    Type  string  `json:"type"`
    Chance float64 `json:"chance"`
}

// Config defines the structure of the configuration JSON file
type Config struct {
    Seed                 int64                `json:"seed"`
    Alpha                float64              `json:"alpha"`
    Beta                 float64              `json:"beta"`
    Damping              float64              `json:"damping"`
    WeekendDamping       float64              `json:"weekend-damping"`
    WeekendDampingOffset int                  `json:"weekend-damping-offset"`
    WeekendDampingScale  int                  `json:"weekend-damping-scale"`
    SessionGap           int                  `json:"session-gap"`
    StartDate            string               `json:"start-date"`
    EndDate              string               `json:"end-date"`
    NUsers               int                  `json:"n-users"`
    FirstUserID          int                  `json:"first-user-id"`
    GrowthRate           float64              `json:"growth-rate"`
    Tag                  string               `json:"tag"`
    ContentTypes         []ContentType        `json:"content-types"`
    AdConfig             AdConfig             `json:"ad-config"`
    Genres               []Preference         `json:"genres"`
    Shows                []Preference         `json:"shows"`
    AuthLevels           []Preference         `json:"auth-levels"`
    SubscriptionChances  []SubscriptionChance `json:"subscription-chances"`
}


// {
//     "seed": 1,
//     "alpha": 90.0,
//     "beta": 604800.0,
//     "damping": 0.09375,
//     "weekend-damping": 0.5,
//     "weekend-damping-offset": 180,
//     "weekend-damping-scale": 360,
//     "session-gap": 1800,
//     "start-date": "2024-04-18T00:00:00Z",
//     "end-date": "2024-04-25T00:00:00Z",
//     "n-users": 5000,
//     "first-user-id": 1,
//     "growth-rate": 0.05,
//     "tag": "streaming-simulation",
//     "genres": [
//         {"name": "Action", "weight": 20},
//         {"name": "Drama", "weight": 30},
//         {"name": "Comedy", "weight": 50}
//     ],
//     "shows": [
//         {"name": "Show 1", "weight": 10},
//         {"name": "Show 2", "weight": 5},
//         {"name": "Show 3", "weight": 85}
//     ],
//     "auth-levels": [
//         {"name": "Guest", "weight": 30},
//         {"name": "Logged In", "weight": 60},
//         {"name": "Logged Out", "weight": 10}
//     ],
//     "subscription-chances": [
//         {"type": "Free", "chance": 0.5},
//         {"type": "Basic", "chance": 0.3},
//         {"type": "Premium", "chance": 0.2}
//     ]
// }

// LoadConfig reads and parses the configuration file into a Config struct
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path) 
	if err != nil {
			return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
			return nil, err
	}
	return &config, nil
}