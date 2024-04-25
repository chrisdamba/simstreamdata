package config

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// ContentType defines the types of content and their distribution weights
type ContentType struct {
	Type   string `mapstructure:"type"`
	Weight int    `mapstructure:"weight"`
}

// AdEvent defines the structure for ad-related events and their weights
type AdEvent struct {
	Event  string `mapstructure:"event"`
	Weight int    `mapstructure:"weight"`
}

// AdConfig holds the configuration for advertisement behavior in audio and video streams
type AdConfig struct {
	AudioAdFrequency  float64   `mapstructure:"audio-ad-frequency"`
	VideoAdFrequency  float64   `mapstructure:"video-ad-frequency"`
	AdEvents          []AdEvent `mapstructure:"ad-events"` 
	PreRollFrequency  float64   `mapstructure:"pre-roll-ad-frequency"` 
	PreRollCooldown   time.Duration `mapstructure:"pre-roll-ad-cooldown"`  
	MidRollWindow     time.Duration `mapstructure:"mid-roll-ad-window"` 
}

type Transition struct {
	Source StateConfig `mapstructure:"source"`
	Dest   StateConfig `mapstructure:"dest"`
	P      float64     `mapstructure:"p"`  // Probability of this transition
}
// SessionPage defines the configuration for different pages that can be accessed in a session.
type SessionPage struct {
	Page    string `mapstructure:"page"`
	Method  string `mapstructure:"method"`
	Status  int    `mapstructure:"status"`
	Auth    string `mapstructure:"auth"`
	Level   string `mapstructure:"level"`
	Weight  int    `mapstructure:"weight"`
}
// {
    // "ad-config": {
    //   "audio-ad-frequency": 0.2, // 20% chance of audio ad
    //   "video-ad-frequency": 0.4, 
    //   "ad-events": [ /* ... your AdEvent definitions ... */ ],
    //   "pre-roll-ad-frequency": 0.6, 
    //   "pre-roll-ad-cooldown": 60 * time.Second, // 1 minute cooldown
    //   "mid-roll-ad-window": 30 * time.Second  // 30 seconds for mid-roll
    // }
//   }

type StateConfig struct {
	Page   string `mapstructure:"page"`
	Method string `mapstructure:"method"`
	Status int    `mapstructure:"status"`
	Auth   string `mapstructure:"auth"`
	Level  string `mapstructure:"level"`
}

type Preference struct {
	Name   string `mapstructure:"name"`
	Weight int    `mapstructure:"weight"`
}

type SubscriptionChance struct {
	Type  string  `mapstructure:"type"`
	Chance float64 `mapstructure:"chance"`
}

type Video struct {
	ID             string
	TitleType      string
	PrimaryTitle   string
	OriginalTitle  string
	IsAdult        bool
	StartYear      string
	EndYear        string
	RuntimeMinutes time.Duration
	Genres         []string
}

type Config struct {
	Seed                 int64                `mapstructure:"seed"`
	Alpha                float64              `mapstructure:"alpha"`
	Beta                 float64              `mapstructure:"beta"`
	Damping              float64              `mapstructure:"damping"`
	WeekendDamping       float64              `mapstructure:"weekend-damping"`
	WeekendDampingOffset int                  `mapstructure:"weekend-damping-offset"`
	WeekendDampingScale  int                  `mapstructure:"weekend-damping-scale"`
	SessionGap           int                  `mapstructure:"session-gap"`
	StartDate            string               `mapstructure:"start-date"`
	EndDate              string               `mapstructure:"end-date"`
	NUsers               int                  `mapstructure:"n-users"`
	FirstUserID          int                  `mapstructure:"first-user-id"`
	GrowthRate           float64              `mapstructure:"growth-rate"`
	Tag                  string               `mapstructure:"tag"`
	ContentTypes         []ContentType        `mapstructure:"content-types"`
	AdConfig             AdConfig             `mapstructure:"ad-config"`
	Genres               []Preference         `mapstructure:"genres"`
	Shows                []Preference         `mapstructure:"shows"`
	Levels               []Preference         `mapstructure:"levels"`
	AuthLevels           []Preference         `mapstructure:"auth-levels"`
	SubscriptionChances  []SubscriptionChance `mapstructure:"subscription-chances"`
	NewSessionPages      []SessionPage    		`mapstructure:"new-session"`
	Transitions      		 []Transition 				`mapstructure:"transitions"`	
	AllVideos 					 []Video 							`mapstructure:"all-videos"` // List of all videos loaded from IMDb

	SimulateVideo     		bool          			`mapstructure:"simulate-video"` 
	AttritionRate     		float64       			`mapstructure:"attrition-rate"`
	StartTime         		time.Time     			`mapstructure:"start-time"` 
	EndTime           		time.Time     			`mapstructure:"end-time"`
	KafkaEnabled     			bool          			`mapstructure:"kafka-enabled"` 
	KafkaBrokerList   		string        			`mapstructure:"kafka-broker-list"`
	KafkaTopic        		string        			`mapstructure:"kafka-topic"`
	OutputFile        		string        			`mapstructure:"output-file"`
	Continuous        		bool          			`mapstructure:"continuous"` 
}

// LoadConfig initializes and reads the configuration using Viper
func LoadConfig(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Default config location
		viper.AddConfigPath("configs")
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv() // Read in environment variables that match

	// Set default for start time as the current time if not provided
	viper.SetDefault("start-time", time.Now().Format(time.RFC3339))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	decoderConfigOption := viper.DecoderConfigOption(func(config *mapstructure.DecoderConfig) {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
				config.DecodeHook,
				mapstructure.StringToTimeHookFunc(time.RFC3339), 
		)
	})
	if err := viper.Unmarshal(&config, decoderConfigOption); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	return &config, nil
}

func LoadVideosFromIMDb(filename string) ([]Video, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    gz, err := gzip.NewReader(file)
    if err != nil {
        return nil, err
    }
    defer gz.Close()

    scanner := bufio.NewScanner(gz)
    videos := []Video{}

    // Skip header line
    scanner.Scan()

    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, "\t")

        runtimeMinutes, _ := strconv.Atoi(parts[7])
        genres := strings.Split(parts[8], ",")

        video := Video{
            ID:             parts[0],
            TitleType:      parts[1],
            PrimaryTitle:   parts[2],
            OriginalTitle:  parts[3],
            IsAdult:        parts[4] == "1",
            StartYear:      parts[5],
            EndYear:        parts[6],
            RuntimeMinutes: time.Duration(runtimeMinutes) * time.Minute,
            Genres:         genres,
        }
        videos = append(videos, video)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return videos, nil
}


// Add a method to the Config struct to load videos into it
func (c *Config) InitializeVideos(filePath string) error {
	videos, err := LoadVideosFromIMDb(filePath)
	if err != nil {
			return err
	}
	c.AllVideos = videos
	return nil
}