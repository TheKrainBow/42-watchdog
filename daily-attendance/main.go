package main

import (
	"fmt"
	"os"
	"time"
	"watchdog/config"
	"watchdog/watchdog"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Invalid usage:")
		fmt.Println("./watchdog <path_to_config> <path_to_log_folder> [YYYY-MM-DD]")
		os.Exit(1)
	}

	// Validate log path
	logPath := os.Args[2]
	info, err := os.Stat(logPath)
	if err != nil {
		fmt.Printf("Invalid path '%s'\n", logPath)
		os.Exit(1)
	} else if !info.IsDir() || logPath[len(logPath)-1] != '/' {
		fmt.Printf("Path '%s' must be a directory and must end with '/'\n", logPath)
		os.Exit(1)
	}

	// Determine target date
	var targetDate time.Time
	if len(os.Args) >= 4 {
		targetDate, err = time.Parse("2006-01-02", os.Args[3])
		if err != nil {
			fmt.Printf("Invalid date format: '%s'. Expected format: YYYY-MM-DD\n", os.Args[3])
			os.Exit(1)
		}
	} else {
		targetDate = time.Now()
	}

	// Init
	config.LoadConfig(os.Args[1])
	logFile := fmt.Sprintf("%swatchdog-%s.log", logPath, targetDate.Format("2006-01-02"))
	err = watchdog.Init(logFile)
	if err != nil {
		fmt.Printf("Error initializing logs: %s\n", err.Error())
		return
	}

	watchdog.Watch(targetDate)
	watchdog.CloseLogs()
}
