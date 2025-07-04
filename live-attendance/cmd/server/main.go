package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"watchdog/config"
	"watchdog/watchdog"
)

func startHTTPServer(port string) {
	http.Handle("/webhook/access-control", verifySignatureMiddleware(http.HandlerFunc(accessControlEndpoint)))
	http.HandleFunc("/commands", commandHandler)

	watchdog.Log(fmt.Sprintf("[HTTP] Listening on port %s", port))
	watchdog.Log("[HTTP] ┌─ Available endpoints:")
	watchdog.Log("       ├── /commands")
	watchdog.Log("       └── /webhook/access-control")

	watchdog.AllowEvents(true)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		watchdog.Log(fmt.Sprintf("[HTTP] [FATAL] could not start server: %s\n", err))
		os.Exit(1)
	}
}

func main() {
	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	if len(os.Args) <= 2 {
		fmt.Printf("Invalid program usage:\n")
		fmt.Printf("./watchdog <path_to_config_file> <path_to_log_file>\n")
		os.Exit(1)
	}

	configFile := os.Args[1]
	logFile := os.Args[2]

	info, err := os.Stat(logFile)
	if err == nil && info.IsDir() {
		fmt.Printf("path '%s' must be a file, not a directory\n", logFile)
		os.Exit(1)
	}

	err = watchdog.InitLogs(logFile)
	if err != nil {
		fmt.Printf("ERROR: couldn't init logs")
		os.Exit(1)
	}
	watchdog.Log(fmt.Sprintf("[WATCHDOG] 📝 Initialiazed log file %s", logFile))
	watchdog.Log(fmt.Sprintf("[WATCHDOG] 💾 Loading config using file %s", configFile))
	config.LoadConfig(configFile)
	err = watchdog.InitAPIs()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	watchdog.Init(watchdog.ConfMailer{
		SmtpServer: "smtp-relay.gmail.com",
		SmtpPort:   587,
		SmtpAuth:   false,
		SmtpUser:   "",
		SmtpPass:   "",
		SmtpTls:    true,
		Helo:       "pedago.42nice.fr",
		FromName:   "Watchdog",
		FromMail:   "pedago.watchdog.noreply@42nice.fr",
	})

	go startHTTPServer("8042")

	// Wait a SIGINT or SIGTERM signal to stop
	sig := <-shutdownSignals
	fmt.Printf("\n") // Used to not display log on the same line as ^C
	watchdog.Log(fmt.Sprintf("Received signal: %v. Starting graceful shutdown...", sig))
	watchdog.PostApprenticesAttendances()
	watchdog.AllowEvents(false)
	watchdog.Log("Watchdog shut down successfully")
	watchdog.Log("")
	watchdog.CloseLogs()
}
