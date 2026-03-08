package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// function called at startup to load the config file or create one if it doesn't exist
func readConfig(configFilepath string) map[string]any {
	// check that the config file actually exists first, if not then we fill it in with the default template
	var cm map[string]any

	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		// config file doesn't exist
		configFile, err := os.Create(configFilepath)
		if err != nil {
			log.Panic("Error creating missing config file")
		}
		defer configFile.Close()
	}
	return cm
}

func startup() error {
	defer func() {
		// set up a process for handling panics in startup
		if r := recover(); r != nil {
			log.Fatal("Ran into a critical error and panicked during startup:", r) // log.Fatal automatically calls os.Exit(1)
		}
	}()

	// create the config map for config values
	configMap := readConfig("../panopticon.config")

	// get the local host for logging real quick
	hostname, err := os.Hostname()
	if err != nil {
		return errors.New("Couldn't get hostname")
	}
	fmt.Println("Starting up Panopticon on host", hostname)

	fmt.Println("Beginning database initialization")
	// call our connectToDatabase function to initalize the connection to the SQLite database, using a type assertion to safely pull the database filepath from the config file
	if databaseFilePath, ok := configMap["DATABASE_FILEPATH"].(string); ok {
		db, err := connectToDatabase(databaseFilePath)
		if err != nil {
			log.Panic("Ran into an error connecting to the SQLite database:", err)
		}
	}

	// create a channel to asynchronously collect errors (any kind of error) from the webserver
	webErrors := make(chan any)
	// grab the REST API port from the pre-loaded config file, with a type assertion so Go doesn't yell at me
	if restPort, ok := configMap["REST_PORT"].(int); ok {
		go StartREST(restPort, webErrors) // start our REST API web server with the async channel to collect errors from
	}

	// error collection goroutine, pulls errors from the webserver and reports on any startup panics
	go func() {
		webError := <-webErrors
		if webError != nil {
			fmt.Println("Panic in webserver:", webError)
		}
	}()
	fmt.Println("Startup executed without errors")
	return nil // able to startup without any errors
}

func main() {

	startupErr := startup()
	if startupErr != nil {
		fmt.Println("Fatal error during startup", startupErr)
		os.Exit(1)
	}
	fmt.Println("Locating and pinging watchgroup")
}
