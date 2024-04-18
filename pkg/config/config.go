package config

import (
    "encoding/json"
    "os"
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
    AudioAdFrequency float64   `json:"audio-ad-frequency"`
    VideoAdFrequency float64   `json:"video-ad-frequency"`
    AdEvents         []AdEvent `json:"ad-events"`
}

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