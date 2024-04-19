package simulator

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/chrisdamba/simstreamdata/pkg/config"
	"github.com/chrisdamba/simstreamdata/pkg/models"
)

type OutputDestination interface {
    WriteMessage(msg []byte) error
}

type KafkaOutput struct {
    producer sarama.SyncProducer
    topic    string
}

type Simulator struct {
    Config *config.Config
    Users  []*models.User
}

type FileOutput struct {
    file *os.File
}

type ConsoleOutput struct{}

func (f *FileOutput) WriteMessage(msg []byte) error {
    _, err := f.file.Write(msg)
    return err
}

func (k *KafkaOutput) WriteMessage(msg []byte) error {
    _, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
        Topic: k.topic,
        Value: sarama.ByteEncoder(msg),
    })
    return err
}

func (c *ConsoleOutput) WriteMessage(msg []byte) error {
    _, err := os.Stdout.Write(msg)
    return err
}

func (sim *Simulator) determineOutputDestination(config *config.Config) OutputDestination {
    if config.KafkaEnabled {
        producer, err := sarama.NewSyncProducer(config.KafkaBrokerList, nil) // Assuming 'Brokers' field
        if err != nil {
            log.Fatalf("Failed to create Kafka producer: %s", err)
        }
        // Assuming cleanProducerClose function implemented
        defer cleanProducerClose(producer)

        return &KafkaOutput{producer: producer, topic: config.KafkaTopic} 
    } else if config.OutputFile != "" {
        file, err := os.Create(config.OutputFile)
        if err != nil {
            log.Fatalf("Failed to create output file: %s", err)
        }
        return &FileOutput{file: file}
    }
    return &ConsoleOutput{}
}

// Placeholder: Ensures proper closure of the Kafka producer
func cleanProducerClose(producer sarama.SyncProducer) {
    if err := producer.Close(); err != nil {
        log.Println("Error closing Kafka producer:", err)
    }
}

// Helper to generate log-normal values
func randomLogNormal(mean, stddev float64) float64 {
    return rand.NormFloat64()*stddev + mean
}

func (sim *Simulator) initializeUsers() {
    for i := 0; i < sim.Config.NUsers; i++ {
        // Generate random preferences based on weighted selections
        initialLevel := sim.weightedRandomInitialLevel()

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
            initialLevel,
            subscriptionType,
        )

        sim.AddSession(user)
    }
}

func (sim *Simulator) AddSession(user *models.User) {
    sim.Users = append(sim.Users, user)
}

func (sim *Simulator) weightedRandomAuthLevel() string {
    return sim.selectRandomPreference(sim.Config.AuthLevels).Name
}

func (sim *Simulator) weightedRandomInitialLevel() string {
    return sim.selectRandomPreference(sim.Config.Levels).Name
}

func (sim *Simulator) weightedRandomSubscriptionType() models.SubscriptionType {
    chosen := sim.selectRandomPreference(convertToPreferences(sim.Config.SubscriptionChances))
    return models.SubscriptionType(chosen.Name)
}

func convertToPreferences(subscriptionChances []config.SubscriptionChance) []config.Preference {
    preferences := make([]config.Preference, len(subscriptionChances))
    for i, chance := range subscriptionChances {
        preferences[i] = config.Preference(chance)
    }
    return preferences
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

func (sim *Simulator) selectRandomPreferences(items []config.Preference, count int) []string {
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

func (sim *Simulator) RunSimulation() {
    sim.initializeUsers() // Initialize users as before

    // Ticker for simulation
    ticker := time.NewTicker(1 * time.Second) 
    defer ticker.Stop()

    for range ticker.C {
        output := sim.determineOutputDestination(sim.Config) // Get output destination once

        for _, user := range sim.Users {
            user.NextEvent(sim.Config, output) // Let users generate events
        }
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


func newKafkaProducer(brokerList []string) (sarama.SyncProducer, func(), error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll          // Ensure data is written to all replicas
    config.Producer.Retry.Max = 10                            // Retry up to 10 times to produce messages
    config.Producer.Return.Successes = true

    producer, err := sarama.NewSyncProducer(brokerList, config)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to start Sarama producer: %w", err)
    }

    // Cleanup function to close the producer
    cleanup := func() {
        if err := producer.Close(); err != nil {
            log.Printf("Failed to close Kafka producer: %s", err)
        }
    }

    return producer, cleanup, nil
}
