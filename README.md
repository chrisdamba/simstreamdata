# SimStreamData

A Go-based simulator for generating realistic streaming media usage data, facilitating the analysis and modeling of user behavior on platforms like Spotify and Netflix. It generates event data to help developers and testers simulate and analyse user behaviour and system performance under various conditions. The simulation encompasses aspects of both audio and video streaming, including interactions like play, pause, and advertisement events.

## Key Features

* **Behavioral Modeling:** Captures the nuances of user interaction:
   * Content selection based on individual preferences (audio vs. video, genres).
   * Session start/end patterns influenced by damping factors and time of day.
   * Transitions between states (browsing, playing content, pausing, account activity).

* **OTT Streaming Simulation:**  Realistically models video streaming behaviors:
   * Content of varying lengths (movies, shows, songs).
   * Ad campaign insertions with configurable frequency, type, and rules.
   * Simulation of user reactions to ads (skipping, etc.).

* **Data Generation:** 
   * Serializes events in formats  suitable for downstream analysis and machine learning (CSV, JSON, potentially Avro).
   * Exports data directly to Kafka topics for streaming analytics use cases.

## Project Goals

* **Understand User Behavior:** Aid in analyzing patterns of content consumption, churn prediction, and the effectiveness of UI/UX changes.
* **Test Recommendation Systems:** Provide a realistic data source to evaluate and fine-tune recommendation algorithms.
* **Optimize Ad Strategies:** Experiment with various ad targeting, placement, and frequency strategies to assess their impact on revenue and user experience.

## How Simulation Works

The `SimStreamData` simulator is designed to mimic user interactions with a digital streaming service by generating pseudo-random events based on defined configurations. Here's a breakdown of how the simulation process works:

### 1. Configuration and Initialization
The simulation starts by reading parameters from a configuration file (`config.json`). This file includes settings such as the number of users, the types of content (audio or video), the probability of different events occurring (like playing a song or starting a video), and ad insertion rules.

### 2. User and Session Creation
For each simulated user:
- A session is created with randomly determined characteristics based on the configuration settings.
- The session includes details such as the start time, the type of content (audio or video), and specific user behaviors like engagement levels and subscription tier.

### 3. Event Generation Cycle
Each session enters a cycle of event generation:
- **Content Playback**: Depending on the type of content, sessions might begin with playing audio or video. For video content, the system considers additional complexities like ad insertion points.
- **Ad Insertions**: If the session is for video content, ads may be inserted based on the likelihood defined in the configuration. This can include pre-roll, mid-roll, and post-roll ads, influenced by user attributes and session details.
- **User Interactions**: Throughout the session, user interactions such as pausing, resuming, and stopping the playback are simulated based on probabilistic models that consider user behavior and session characteristics.

### 4. State Management and Transitions
The system maintains the state of each session, updating it with every event based on several factors:
- **Temporal Factors**: Time-driven events are handled, such as transitioning from playing content to inserting an ad after a specific duration.
- **User Behavior**: Randomized decisions simulate user interactions, influencing the flow of the session (e.g., a user might skip an ad or a song).
- **Adherence to Rules**: All session activities adhere to predefined rules set in the configuration, such as ad frequency limits and user subscription privileges.

### 5. Output and Logging
All events and state transitions within each session are logged:
- Detailed logs capture every action within a session, providing insights into the simulated behavior and the effectiveness of different configurations.
- The output can be used for analysis to understand user engagement, system performance, and potential areas for optimization in content delivery and ad placement strategies.

This simulation framework provides a robust tool for developers, testers, and researchers to study and optimize streaming platforms by closely replicating real-world user behavior and system responses under controlled, yet realistic, conditions.

## Usage

1. **Configuration:** Customize the `config.json` file to tailor the simulation:
   * User attributes (engagement levels, subscription tiers).
   * Content preference weights.
   * State transition probabilities.
   * Ad settings.

2. **Run the Simulation:**
   `go run main.go <config_file_path>`

3. **Output:** The simulator will generate event data in the specified format, ready for analysis.

## Configuration Example (config.json)

```json
{
  "seed": 12345,
  "alpha": 120.0,
  "beta": 3600.0,
  // ... other configuration settings ...
  "content-types": [
    {"type": "audio", "weight": 50},
    {"type": "video", "weight": 50}
  ],
  "ad-config": {
    // ... ad-related settings ...
  }
}
