package cmd

import (
    "fmt"
    "os"
    "time"

    "github.com/spf13/cobra"
		"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "simstreamdata",
    Short: "Simulates streaming data for media platforms",
    Long: `Simstreamdata is a CLI tool to simulate event streaming data from users 
interacting with a media platform like video or music streaming services.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Your simulation entry point
        fmt.Println("Simulation started...")
        simulate()
    },
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.simstreamdata.yaml)")
    rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    rootCmd.Flags().Int("nusers", 1000, "initial number of users")
    rootCmd.Flags().Float64("growth-rate", 0.01, "annual user growth rate (as a fraction, e.g., 1% => 0.01)")
    rootCmd.Flags().String("start-time", time.Now().Format(time.RFC3339), "start time for data generation")
    rootCmd.Flags().String("end-time", time.Now().AddDate(0, 0, 7).Format(time.RFC3339), "end time for data generation")
    rootCmd.Flags().String("kafka-broker", "", "kafka broker list")
    rootCmd.Flags().String("kafka-topic", "", "kafka topic for sending data")
    rootCmd.Flags().String("output-file", "", "file to write output to (default is stdout)")
    rootCmd.Flags().Bool("realtime", false, "run simulation in real-time")
    // Add more flags as required
}

func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag.
    } else {
        // Find home directory.
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        // Search config in home directory with name ".cobra" (without extension).
        viper.AddConfigPath(home)
        viper.SetConfigType("yaml")
        viper.SetConfigName(".simstreamdata")
    }

    viper.AutomaticEnv() // read in environment variables that match

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    }
}

// Execute executes the root command.
func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}
