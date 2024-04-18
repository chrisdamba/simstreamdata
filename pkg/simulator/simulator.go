package simulator

import (
    "math/rand"

		"github.com/chrisdamba/simstreamdata/pkg/config"
		"github.com/chrisdamba/simstreamdata/pkg/models"
)

type Simulator struct {
    Config *config.Config
    Users  []*models.User
}

func NewSimulator(cfg *config.Config) *Simulator {
    rand.Seed(cfg.Seed)
    return &Simulator{
        Config: cfg,
        Users:  make([]*models.User, cfg.NUsers),
    }
}

func (sim *Simulator) initializeUsers() {
    for i := 0; i < sim.Config.NUsers; i++ {
        // Generate random preferences based on weighted selections
        preferredGenres := sim.selectRandomPreferences(sim.Config.Genres, 3)
        favoriteShows := sim.selectRandomPreferences(sim.Config.Shows, 3)

        // Determine the authorization level and subscription type with weights
        authLevel := sim.weightedRandomAuthLevel()
        subscriptionType := sim.weightedRandomSubscriptionType()

        // Create new user
        startTime := sim.Config.StartTime.Add(time.Duration(i) * time.Minute)
        user := models.NewUser(
            randomLogNormal(sim.Config.Alpha, 0.5),
            randomLogNormal(sim.Config.Beta, 0.5),
            startTime,
            authLevel,
            subscriptionType,
            preferredGenres,
            favoriteShows,
            sim.randomViewingHours(),
        )

        sim.AddSession(user)
    }
}


func (sim *Simulator) weightedRandomAuthLevel() string {
    return sim.selectRandomPreference(sim.Config.AuthLevels).Name
}

func (sim *Simulator) weightedRandomSubscriptionType() models.SubscriptionType {
    chosen := sim.selectRandomPreference(sim.Config.SubscriptionChances)
    return models.SubscriptionType(chosen.Name)
}

func (sim *Simulator) selectRandomPreference(preferences []config.Preference) config.Preference {
    totalWeight := 0
    for _, p := range preferences {
        totalWeight += p.Weight
    }
    r := rand.Intn(totalWeight)
    for _, p := range preferences {
        if r < p.Weight {
            return p
        }
        r -= p.Weight
    }
    return preferences[0] // default fallback
}

func (sim *Simulator) randomViewingHours() int {
    return rand.Intn(41) // Random hours from 0 to 40
}

func (sim *Simulator) selectRandomPreferences(items []config.PreferenceItem, count int) []string {
    selected := make([]string, count)
    for i := 0; i < count; i++ {
        totalWeight := 0
        for _, item := range items {
            totalWeight += item.Weight
        }
        r := rand.Intn(totalWeight)
        for _, item := range items {
            if r < item.Weight {
                selected[i] = item.Name
                break
            }
            r -= item.Weight
        }
    }
    return selected
}
func (s *Simulator) RunSimulation() {
    for _, _ := range s.Users {
        // Determine if the session involves video and possibly trigger ads
        if contentType := s.pickContentType(); contentType == "video" {
            if rand.Float64() < s.Config.AdConfig.VideoAdFrequency {
                s.handleVideoAds()
            }
        }
        // Continue with other simulation tasks
    }
}

func (s *Simulator) pickContentType() string {
    totalWeight := 0
    for _, ct := range s.Config.ContentTypes {
        totalWeight += ct.Weight
    }
    r := rand.Intn(totalWeight)
    sum := 0
    for _, ct := range s.Config.ContentTypes {
        sum += ct.Weight
        if r < sum {
            return ct.Type
        }
    }
    return "audio" // default if something goes wrong
}

func (s *Simulator) handleVideoAds() {
    // Process video ads based on configuration
    for _, ad := range s.Config.AdConfig.AdEvents {
        if rand.Float64() < float64(ad.Weight) {
            // Log or handle ad event
        }
    }
}


func (sim *Simulator) determineAuthLevel() string {
    // Randomly determine auth level; this is simplified, expand as needed
    authLevels := []string{"Guest", "Logged In", "Logged Out"}
    return authLevels[rand.Intn(len(authLevels))]
}
