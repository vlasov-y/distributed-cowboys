// ┌─┐┌─┐┌┐┐┬─┐o┌─┐
// │  │ ││││├─ ││ ┬
// └─┘┘─┘┘└┘┘  ┘┘─┘

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type OperationMode int

const (
	Seed OperationMode = iota
	Regular
)

type Configuration struct {
	WorkerName               string        `yaml:"workerName"`
	DatabaseConnectionString string        `yaml:"databaseConnectionString"`
	ServerPort               uint16        `yaml:"serverPort"`
	ShootingDelay            uint16        `yaml:"shootingDelay"`
	OperationMode            OperationMode `yaml:"operationMode"`
	GenerateRandomCowboys    uint16        `yaml:"generateRandomCowboys"`
}

func LoadConfiguration() (*Configuration, error) {
	// Load extra .env if present
	godotenv.Load(".env")

	var configuration Configuration
	// Select operation mode
	if strings.ToLower(os.Getenv("OPERATION_MODE")) == "seed" {
		configuration.OperationMode = Seed
		if num, err := strconv.ParseUint(os.Getenv("GENERATE_RANDOM_COWBOYS"), 10, 16); err == nil {
			log.Printf("Will create %d new cowboys\n", num)
			configuration.GenerateRandomCowboys = uint16(num)
		}
	} else {
		configuration.OperationMode = Regular
		configuration.GenerateRandomCowboys = 0
	}
	// Load configuration from YAML
	if err := yaml.Unmarshal([]byte(os.Getenv("CONFIGURATION")), &configuration); err != nil {
		message := fmt.Sprintf("Failed to unmarshal the configuration: %s", err.Error())
		return nil, errors.New(message)
	}
	// Set worker name to the hostname if empty or not overridden
	if len(configuration.WorkerName) == 0 {
		var err error
		if configuration.WorkerName, err = os.Hostname(); err != nil {
			log.Panicf("Could not get hostname: %v\n", err)
		}
	}
	// Override custom server port
	if num, err := strconv.ParseUint(os.Getenv("SERVER_PORT"), 10, 16); err == nil {
		configuration.ServerPort = uint16(num)
	}
	return &configuration, nil
}
