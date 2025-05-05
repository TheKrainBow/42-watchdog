package main

import (
	"fmt"
	"os"
	"time"
	"watchdog/config"
	"watchdog/watchdog"
)

func main() {
	today := time.Now()
	if len(os.Args) <= 2 {
		fmt.Printf("Invalid program usage:\n")
		fmt.Printf("./watchdog <path_to_config> <path_to_log_folder>\n")
		os.Exit(1)
	}
	info, err := os.Stat(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid path '%s'\n", os.Args[2])
		os.Exit(1)
	} else if !info.IsDir() || os.Args[2][len(os.Args[2])-1] != '/' {
		fmt.Printf("path '%s' must be a directory, and not a file\n", os.Args[2])
		os.Exit(1)
	}
	config.LoadConfig(os.Args[1])
	err = watchdog.Init(fmt.Sprintf("%swatchdog-%s.log", os.Args[2], today.Format("2006-01-02")))
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
	watchdog.Watch(today)
	watchdog.CloseLogs()
}
