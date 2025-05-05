package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type CommandRequest struct {
	Command    string                 `json:"command"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

var serverURL string

func sendCommand(command string, params map[string]interface{}) {
	cmdReq := CommandRequest{
		Command:    command,
		Parameters: params,
	}

	jsonBody, err := json.Marshal(cmdReq)
	if err != nil {
		fmt.Println("JSON encode error:", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Request error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	fmt.Printf("Response: %+v\n", respBody)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "watchdog-client",
		Short: "Client for sending commands to the watchdog server",
	}

	rootCmd.PersistentFlags().StringVarP(&serverURL, "url", "u", "http://localhost:8080/commands", "Full URL of the server endpoint")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Send start_listen command",
		Run: func(cmd *cobra.Command, args []string) {
			sendCommand("start_listen", nil)
		},
	})

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Send stop_listen command",
		Run: func(cmd *cobra.Command, args []string) {
			postAttendance, _ := cmd.Flags().GetBool("post-attendance")
			params := map[string]interface{}{}
			if postAttendance {
				params["post_attendance"] = true
			}
			sendCommand("stop_listen", params)
		},
	}
	stopCmd.Flags().Bool("post-attendance", false, "Post attendance after stopping")
	rootCmd.AddCommand(stopCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Send get_status command",
		Run: func(cmd *cobra.Command, args []string) {
			sendCommand("get_status", nil)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "notify",
		Short: "Send notify_students command",
		Run: func(cmd *cobra.Command, args []string) {
			sendCommand("notify_students", nil)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
