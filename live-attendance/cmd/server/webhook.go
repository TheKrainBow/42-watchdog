package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"watchdog/watchdog"
)

// Handler for the /command endpoint
func commandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add Authentication/Authorization here!
	// This endpoint should be protected (e.g., API key, Basic Auth, IP restriction)
	// Example: Check for a specific header/token
	// apiKey := r.Header.Get("X-Admin-Token")
	// if apiKey != "your-secure-admin-token" {
	//     http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//     return
	// }

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var cmdReq CommandRequest
	err = json.Unmarshal(bodyBytes, &cmdReq)
	if err != nil {
		http.Error(w, "Invalid command format (expecting JSON with 'command' field)", http.StatusBadRequest)
		return
	}

	watchdog.Log(fmt.Sprintf("[CLI] 🛠️  Received command: %s", cmdReq.Command))
	responseMessage := ""
	statusCode := http.StatusOK
	// Process the command
	switch cmdReq.Command {
	case "start_listen":
		watchdog.AllowEvents(true)
		responseMessage = "Enabled listening hooks (Check server logs for more details)"
	case "stop_listen":
		shouldPost := false
		if params := cmdReq.Parameters; params != nil {
			if postVal, ok := params["post_attendance"]; ok {
				if postBool, ok := postVal.(bool); ok {
					shouldPost = postBool
					watchdog.Log(fmt.Sprintf("[CLI]       With argument post_attendance: %t", shouldPost))
				}
			}
		}
		watchdog.AllowEvents(false)
		if shouldPost {
			watchdog.PostApprenticesAttendances()
			responseMessage = "Disabled listening hooks and posted attendances (Check server logs for more details)"
		} else {
			responseMessage = "Disabled listening hooks (Check server logs for more details)"
		}
	case "get_status":
		watchdog.PrintUsersTimers()
		responseMessage = "Check server logs for status detail"
	case "notify_students":
		statusCode = http.StatusNotImplemented
		responseMessage = "Coming soon"
		return
	default:
		responseMessage = fmt.Sprintf("Unknown command: %s", cmdReq.Command)
		statusCode = http.StatusBadRequest
	}

	// Send response
	w.WriteHeader(statusCode)
	fmt.Fprint(w, responseMessage)
}

// Middleware function to verify the webhook signature
func verifySignatureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			watchdog.Log(fmt.Sprintf("Middleware: Method not allowed: %s", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		receivedSigHex := r.Header.Get("x-webhook-signature")
		if receivedSigHex == "" {
			watchdog.Log("Middleware: Missing x-webhook-signature header")
			http.Error(w, "Missing signature header", http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			watchdog.Log(fmt.Sprintf("Middleware: Error reading request body: %v", err))
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// Restore the body so the next handler can read it.
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		mac := hmac.New(sha512.New, []byte(webhookSecretKey))
		mac.Write(bodyBytes)
		expectedSigBytes := mac.Sum(nil)
		calculatedSigHex := hex.EncodeToString(expectedSigBytes)

		if !hmac.Equal([]byte(calculatedSigHex), []byte(receivedSigHex)) {
			watchdog.Log("Middleware: Invalid signature. Request rejected")
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func accessControlEndpoint(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Handler: Error reading restored request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload CAPayload
	err = json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		log.Printf("Handler: Error unmarshalling webhook JSON: %v", err)
		http.Error(w, "Invalid payload format", http.StatusBadRequest)
		return
	}

	// Only allow events that are "Access Granted" types
	if payload.Data.Code != 48 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Webhook received, event ignored (code != 48)")
		return
	}

	if payload.Data.User == nil {
		log.Printf("Handler: User ID is null")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Webhook received with empty user")
		return
	}
	layout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Europe/Paris")
	eventTime, err := time.ParseInLocation(layout, payload.Data.DateTime, loc)
	if err != nil {
		log.Printf("Handler: Error parsing event time '%s': %v", payload.Data.DateTime, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Webhook couldn't parse event time")
	}
	go watchdog.UpdateUserAccess(*payload.Data.User, payload.Data.Event.UserName, eventTime, payload.Data.Event.DoorName)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook received and queued to process")
}
