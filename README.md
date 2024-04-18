# SimStreamData

A Go-based simulator for generating realistic streaming media usage data, facilitating the analysis and modeling of user behavior on platforms like Spotify and Netflix.

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
