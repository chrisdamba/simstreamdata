package cmd

import (
	"fmt"
	"os"
	"reflect"
	"time"

	// "time"

	"github.com/chrisdamba/simstreamdata/pkg/config"
    "github.com/chrisdamba/simstreamdata/pkg/simulator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "simstreamdata",
    Short: "Simulates streaming data for media platforms",
    Long: `Simstreamdata is a CLI tool to simulate event streaming data from users interacting with a media platform like video or music streaming services.`,
    Run: func(cmd *cobra.Command, args []string) {
        now := time.Now().Format(time.RFC3339)
        viper.SetDefault("start-time", now)  // This sets it if not already set via config or flags

        cfg, err := config.LoadConfig(cfgFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
            os.Exit(1)
        }

        fmt.Println("Simulation started with the following configuration:")
        v := reflect.ValueOf(cfg).Elem()
        t := v.Type()
        for i := 0; i < v.NumField(); i++ {
            field := v.Field(i)
            fmt.Printf("%s: %v\n", t.Field(i).Name, field.Interface())
        }
        sim := simulator.NewSimulator(cfg)
        sim.RunSimulation()
    },
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.simstreamdata.yaml)")
    // rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    // rootCmd.Flags().Int("nusers", 1000, "initial number of users")
    // rootCmd.Flags().Float64("growth-rate", 0.01, "annual user growth rate (as a fraction, e.g., 1% => 0.01)")
    // rootCmd.Flags().String("start-time", time.Now().Format(time.RFC3339), "start time for data generation")
    // rootCmd.Flags().String("end-time", time.Now().AddDate(0, 0, 7).Format(time.RFC3339), "end time for data generation")
    // rootCmd.Flags().String("kafka-broker", "", "kafka broker list")
    // rootCmd.Flags().String("kafka-topic", "", "kafka topic for sending data")
    // rootCmd.Flags().String("output-file", "", "file to write output to (default is stdout)")
    // rootCmd.Flags().Bool("realtime", false, "run simulation in real-time")
    // Add more flags as required

    rootCmd.Flags().Float64("alpha", 0.0, "Alpha value for simulation")
	rootCmd.Flags().Float64("beta", 0.0, "Beta value for simulation")
    rootCmd.Flags().String("kafka-broker", "", "kafka broker list")
    rootCmd.Flags().Int("n-users", 1000, "initial number of users")
    rootCmd.Flags().Bool("simulate-video", false, "simulate video streaming events")
    rootCmd.Flags().Float64("growth-rate", 0.0, "annual user growth rate")
    rootCmd.Flags().Float64("attrition-rate", 0.0, "annual user attrition rate")
    rootCmd.Flags().String("start-time", time.Now().Format(time.RFC3339), "start time for data generation")
    rootCmd.Flags().String("end-time", time.Now().AddDate(0, 0, 7).Format(time.RFC3339), "end time for data generation")
    rootCmd.Flags().Bool("kafka-enabled", false, "is kafka enabled")
    rootCmd.Flags().String("kafka-broker-list", "", "kafka broker list")
    rootCmd.Flags().String("kafka-topic", "", "kafka topic for sending data")
    rootCmd.Flags().String("output-file", "", "file to write output to (default is stdout)")
    rootCmd.Flags().Bool("continuous", false, "run simulation in real-time") 

	viper.BindPFlags(rootCmd.Flags())
}

func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag.
        viper.SetConfigFile(cfgFile)
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
