package simulator

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/chrisdamba/simstreamdata/pkg/config"
	"github.com/chrisdamba/simstreamdata/pkg/models"
)

type OutputDestination interface {
    WriteMessage(topic string, msg []byte) error
}

type KafkaOutput struct {
    producer sarama.SyncProducer
}

type Simulator struct {
    Config *config.Config
    Users  []*models.User
    StateMachine *models.StateMachine
}

type FileOutput struct {
    files map[string]*os.File
    basePath string // Base directory for output files
}

// NewFileOutput creates a new FileOutput instance with initialized values.
func NewFileOutput(basePath string) *FileOutput {
    return &FileOutput{
        files: make(map[string]*os.File),
        basePath: basePath,
    }
}

type ConsoleOutput struct{}

func NewSimulator(cfg *config.Config) *Simulator {
    return &Simulator{
        Config: cfg,
        Users:  []*models.User{},
    }
}

func (f *FileOutput) WriteMessage(topic string, msg []byte) error {
    // Check if the file already exists in the map
    if _, ok := f.files[topic]; !ok {
        // If not, create the file
        filename := fmt.Sprintf("%s/%s.txt", f.basePath, topic)
        file, err := os.Create(filename)
        if err != nil {
            return fmt.Errorf("failed to create file for topic %s: %w", topic, err)
        }
        f.files[topic] = file
    }

    // Write the message to the corresponding file
    _, err := f.files[topic].Write(msg)
    if err != nil {
        return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
    }

    return nil
}


func (k *KafkaOutput) WriteMessage(topic string, msg []byte) error {
    if k.producer == nil {
        return fmt.Errorf("Kafka producer is closed")
    }
    _, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.ByteEncoder(msg),
    })
    return err
}


func (c *ConsoleOutput) WriteMessage(topic string, msg []byte) error {
    _, err := os.Stdout.Write(msg)
    return err
}

// Ensure producer is closed properly after all messages are sent
func (sim *Simulator) determineOutputDestination(config *config.Config) OutputDestination {
    if config.KafkaEnabled {
        brokerList := strings.Split(config.KafkaBrokerList, ",")
        producer, err := sarama.NewSyncProducer(brokerList, nil)
        if err != nil {
            log.Fatalf("Failed to create Kafka producer: %s", err)
        }
        return &KafkaOutput{producer: producer}
    } else if config.OutputFile != "" {
        return NewFileOutput(config.OutputFile)
    }
    return &ConsoleOutput{}
}

// Helper to generate log-normal values
func randomLogNormal(mean, stddev float64) float64 {
    return rand.NormFloat64()*stddev + mean
}

func (sim *Simulator) initializeUsers() {
    stateMachine := models.InitializeStates(sim.Config) // Initialize a state machine for all users (or per user if different)
    for i := 0; i < sim.Config.NUsers; i++ {
        // Generate random preferences based on weighted selections
        initialLevel := sim.weightedRandomInitialLevel()

        // Determine the authorization level and subscription type with weights
        authLevel := sim.weightedRandomAuthLevel()
        subscriptionType := sim.weightedRandomSubscriptionType()

        // Generate random genre preferences
        genrePreferences := sim.generateRandomGenrePreferences()

        // Create new user
        startTime := sim.Config.StartTime.Add(time.Duration(i) * time.Minute)
        user := models.NewUser(
            randomLogNormal(sim.Config.Alpha, 0.5),
            randomLogNormal(sim.Config.Beta, 0.5),
            startTime,
            authLevel,
            initialLevel,
            subscriptionType,
            stateMachine,
            genrePreferences,
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

// generateRandomGenrePreferences generates a map of genres with randomized weights based on configured preferences
func (sim *Simulator) generateRandomGenrePreferences() map[string]int {
    genreMap := make(map[string]int)
    for _, genre := range sim.Config.Genres {
        // Randomize genre weight: here we simulate user preference strength by multiplying the base weight by a random factor
        randomFactor := rand.Intn(10) + 1  // Random factor between 1 and 10
        genreMap[genre.Name] = genre.Weight * randomFactor
    }
    return genreMap
}

func convertToPreferences(subscriptionChances []config.SubscriptionChance) []config.Preference {
    preferences := make([]config.Preference, len(subscriptionChances))
    for i, chance := range subscriptionChances {
        // Manually assigning the values and converting Chance to an integer weight if necessary.
        preferences[i] = config.Preference{
            Name: chance.Type,
            Weight: int(chance.Chance * 100), // Assuming Chance is a percentage; adjust the scaling as needed.
        }
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

// Simulation constants
const eventRateCalcInterval = 10 * time.Second

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

func (sim *Simulator) RunSession(user *models.User) {

    // Selecting a video at the start of the session
    selectedVideo := user.SelectVideo(sim.Config)
    user.CurrentSession.CurrentVideo = selectedVideo

    // Simulate watching the video
    fmt.Printf("User %s is watching %s\n", user.ID, selectedVideo.PrimaryTitle)

    // Check video progress to handle end of video or continue playing
    user.CurrentSession.CheckVideoProgress()

    // Simulate the end of the video
    fmt.Printf("User %s finished watching %s\n", user.ID, selectedVideo.PrimaryTitle)

    // Decide next action after the video ends
    if user.DecidesToContinueWatching() {
        // If continuing, select a new video
        selectedVideo = user.SelectVideo(sim.Config)
        user.CurrentSession.CurrentVideo = selectedVideo
        fmt.Printf("User %s continued to watch %s\n", user.ID, selectedVideo.PrimaryTitle)
    } else {
        fmt.Printf("User %s ended their session.\n", user.ID)
    }
}


// RunSimulation starts the simulation process.
func (sim *Simulator) RunSimulation() {
    output := sim.determineOutputDestination(sim.Config)
    defer func() {
        if closer, ok := output.(io.Closer); ok {
            closer.Close()
        }
    }()

    sim.initializeUsers() // Setup initial user base for the simulation.
    log.Printf("Initial number of users: %d\n", sim.Config.NUsers)
    log.Printf("Simulation starts from %s to %s\n", sim.Config.StartTime.UTC().Format(time.RFC3339), sim.Config.EndTime.Format(time.RFC3339))

    // Start the simulation timer
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    // Initialize variables for progress tracking
    var (
        eventsCount    int
        lastReportTime = time.Now().UTC()
    )
        
    // Run the simulation until the current time exceeds the end time
    simulationEndTime, _ := time.Parse(time.RFC3339, sim.Config.EndTime.Format(time.RFC3339))
    for range ticker.C {
        currentUTC := time.Now().UTC()
        if currentUTC.After(simulationEndTime) {
            log.Printf("Simulation end time reached: %s\n", simulationEndTime.Format(time.RFC3339))
            break // Exit the loop to end the simulation
        }
        for _, user := range sim.Users {
            // Ensure that the session exists and is not done
            if user.CurrentSession == nil || user.CurrentSession.IsDone() {
                rng := rand.New(rand.NewSource(currentUTC.UnixNano()))
                user.CurrentSession = models.NewSession(user.ID.String(), user.StateMachine, user.SubscriptionType, 0, user.StartTime, rng, sim.Config)
            }

            // Run a session handling logic
            sim.RunSession(user)

            // Process the next event in the current session
            eventMsg, err := user.NextEvent(rand.New(rand.NewSource(currentUTC.UnixNano())), sim.Config)
            if err != nil {
                log.Printf("Error during event generation: %v", err)
                continue
            }
            if err := output.WriteMessage(eventMsg.Topic, eventMsg.Message); err != nil {
                log.Printf("Failed to write message: %v", err)
            }
            eventsCount++
        }

        // Calculate and display the rate of events
        if currentUTC.Sub(lastReportTime) >= eventRateCalcInterval {
            rate := float64(eventsCount) / currentUTC.Sub(lastReportTime).Seconds()
            log.Printf("Time: %s, Events: %d, Rate: %.2f eps\n", currentUTC.Format(time.RFC3339), eventsCount, rate)
            eventsCount = 0
            lastReportTime = currentUTC
        }
    }
    log.Printf("Simulation completed at %s\n", time.Now().UTC().Format(time.RFC3339))
}
