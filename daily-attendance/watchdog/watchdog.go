package watchdog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	"watchdog/apiManager"
	"watchdog/config"
)

func GetDailyUsers(day time.Time) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 7, 30, 0, 0, time.UTC)
	end := time.Date(day.Year(), day.Month(), day.Day(), 20, 30, 0, 0, time.UTC)
	queryUrl := fmt.Sprintf("/events/?format=datatables&start_date=%s&end_date=%s&length=-1", formatTimeForURL(start), formatTimeForURL(end))
	fmt.Printf("Fetching Control Access events\n")
	Log(fmt.Sprintf("Fetching Control Access events for %s ...\n", day.Format("2006-01-02")))
	resp, err := apiManager.GetClient(config.AccessControl).Get(queryUrl)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}

	var res EventResponse
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}

	// Define the layout for the datetime format
	layout := "2006-01-02 15:04:05"

	Log(fmt.Sprintf("Found %d events\n", len(res.Data)))
	for _, event := range res.Data {
		if event.User != nil { // Check if user is not nil
			// Parse the datetime field
			parsedTime, err := time.Parse(layout, event.DateTime)
			if err != nil {
				Log(fmt.Sprintf("ERROR: %s", err.Error()))
				os.Exit(1)
			}

			// Print the parsed datetime and event details
			user, exist := AllUsers[*event.User]
			if !exist {
				user = User{
					ControlAccessID:   *event.User,
					ControlAccessName: event.Data.UserName,
					FirstAccess:       parsedTime,
					LastAccess:        parsedTime,
					Duration:          parsedTime.Sub(parsedTime),
				}
			}
			if user.FirstAccess.IsZero() || user.FirstAccess.After(parsedTime) {
				user.FirstAccess = parsedTime
				user.Duration = user.LastAccess.Sub(user.FirstAccess)
			}
			if user.LastAccess.IsZero() || user.LastAccess.Before(parsedTime) {
				user.LastAccess = parsedTime
				user.Duration = user.LastAccess.Sub(user.FirstAccess)
			}
			AllUsers[*event.User] = user
		}
	}
	totalSteps = len(AllUsers) * 4
}

func UpdateUserWithAccessControl(controlID int) User {
	resp, err := apiManager.GetClient(config.AccessControl).Get(fmt.Sprintf("/users/%d", controlID))
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}

	var res UserResponse
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}
	user := AllUsers[controlID]
	user.Login42 = res.Properties.Login
	user.ID42 = res.Properties.ID
	AllUsers[controlID] = user
	return user
}

func RemoveUserWithoutLogins() {
	i := 1
	total := len(AllUsers)
	loginActualSteps = 1
	loginTotalSteps = total
	for key, value := range AllUsers {
		box := getBoxChar(i, total)
		user := UpdateUserWithAccessControl(key)
		if user.Login42 == "" {
			user.Login42 = "*No Login*"
			Log(fmt.Sprintf("%s ❌ %s: %s\n", box, value.ControlAccessName, user.Login42))
			totalSteps -= 2
			delete(AllUsers, key)
			i++
			PrintLoading()
			loginActualSteps++
			continue
		}
		if user.ID42 == "" {
			user.Login42 = "*No ID42*"
			Log(fmt.Sprintf("%s ❌ %s: %s\n", box, value.ControlAccessName, user.Login42))
			totalSteps -= 2
			delete(AllUsers, key)
			i++
			PrintLoading()
			loginActualSteps++
			continue
		}
		Log(fmt.Sprintf("%s ✅ Renamed %s to %s\n", box, value.ControlAccessName, user.Login42))
		i++
		PrintLoading()
		loginActualSteps++
	}
}

func PrintUsersTimers() {
	for _, value := range AllUsers {
		Log(fmt.Sprintf(" %s: %s -> %s | Total : %dh%dm%ds\n",
			value.Login42,
			value.FirstAccess.Format("15:04:05"),
			value.LastAccess.Format("15:04:05"),
			int(value.Duration.Hours()),
			int(value.Duration.Minutes())%60,
			int(value.Duration.Seconds())%60,
		))
	}
}

func isProjectOngoing(login string, projectID string) bool {
	resp, err := apiManager.GetClient(config.FTv2).Get(fmt.Sprintf("/users/%s/projects/%s/teams?sort=-created_at", login, projectID))
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		return false
	}

	var res []ProjectResponse
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		os.Exit(1)
	}
	return len(res) >= 1 && res[0].Status == "in_progress"
}

func RemoveNonApprenticeUsers() {
	project := ""
	i := 1
	total := len(AllUsers)
	nonApprenticeActualSteps = 1
	nonApprenticeTotalSteps = total
	for key, value := range AllUsers {
		box := getBoxChar(i, total)
		isApprentice := false
		for _, projectID := range config.ConfigData.ApiV2.ApprenticeProjects {
			if isProjectOngoing(value.Login42, projectID) {
				isApprentice = true
				project = projectID
				break
			}
		}
		if !isApprentice {
			Log(fmt.Sprintf("%s ❌ %s: not an apprentice\n", box, value.Login42))
			delete(AllUsers, key)
			totalSteps -= 1
			i++
			PrintLoading()
			nonApprenticeActualSteps++
			continue
		}
		Log(fmt.Sprintf("%s ✅ %s: has project %s on-going\n", box, value.Login42, project))
		i++
		PrintLoading()
		nonApprenticeActualSteps++
	}
}

type APIAttendance struct {
	Begin_at  string `json:"begin_at"`
	End_at    string `json:"end_at"`
	Source    string `json:"source"`
	Campus_id int    `json:"campus_id"`
	User_id   int    `json:"user_id"`
}

func getBoxChar(i int, len int) string {
	if i == len {
		return "┗"
	} else {
		return "┠"
	}
}

func PostApprenticesAttendances() {
	msg := ""
	i := 1
	total := len(AllUsers)
	postAttendanceActualSteps = 1
	postAttendanceTotalSteps = total
	for _, value := range AllUsers {
		var resp *http.Response
		box := getBoxChar(i, total)
		id42, err := strconv.ParseInt(value.ID42, 10, 64)
		if err != nil {
			msg = fmt.Sprintf("couldn't convert string to int \"%s\"", value.ID42)
		} else {
			_ = id42
			if config.ConfigData.Attendance42.AutoPost {
				resp, err = apiManager.GetClient(config.FTAttendance).Post("/attendances", APIAttendance{
					Begin_at:  value.FirstAccess.UTC().Format(time.RFC3339),
					End_at:    value.LastAccess.UTC().Format(time.RFC3339),
					Source:    "access-control",
					Campus_id: 41,
					User_id:   int(id42),
				})
				if err != nil {
					msg = err.Error()
				} else {
					msg = resp.Status
				}
			} else {
				msg = "AUTOPOST is OFF"
			}
		}
		if err != nil || (config.ConfigData.Attendance42.AutoPost && resp.StatusCode != http.StatusOK) {
			Log(fmt.Sprintf("%s ❌ Posted attendance for %s (%dh%dm%ds): %s\n", box, value.Login42, int(value.Duration.Hours()), int(value.Duration.Minutes())%60, int(value.Duration.Seconds())%60, msg))
		} else {
			Log(fmt.Sprintf("%s ✅ Posted attendance for %s (%dh%dm%ds): %s\n", box, value.Login42, int(value.Duration.Hours()), int(value.Duration.Minutes())%60, int(value.Duration.Seconds())%60, msg))
		}
		i++
		PrintLoading()
		postAttendanceActualSteps++
	}
}

func RemoveTooShortDuration(duration time.Duration) {
	i := 1
	total := len(AllUsers)
	durationActualSteps = 1
	durationTotalSteps = total
	for key, value := range AllUsers {
		if value.FirstAccess == value.LastAccess {
			delete(AllUsers, key)
			totalSteps -= 3
			Log(fmt.Sprintf("%s ❌ Removing %s (Used badge only once)\n", getBoxChar(i, total), value.ControlAccessName))
		} else if value.Duration < duration {
			delete(AllUsers, key)
			totalSteps -= 3
			Log(fmt.Sprintf("%s ❌ Removing %s (%.2f minutes)\n", getBoxChar(i, total), value.ControlAccessName, value.Duration.Minutes()))
		}
		i++
		PrintLoading()
		durationActualSteps++
	}
}

func Watch(day time.Time) {
	Log("Starting Watchdog")
	defer Log("")
	GetDailyUsers(day)
	if len(AllUsers) == 0 {
		Log("⚠️  Every users got removed. Stopping script\n")
		return
	}

	Log(fmt.Sprintf("┏ [Step 1] 🕒 [%d users remaining] Removing durations lass than 10 minutes:\n", len(AllUsers)))
	RemoveTooShortDuration(time.Duration(60 * 10)) // Time inside time.Duration is in seconds
	if len(AllUsers) == 0 {
		Log("⚠️  Every users got removed. Stopping script\n")
		PrintLoading()
		return
	}

	Log(fmt.Sprintf("┏ [Step 2] 📝 [%d users remaining] Removing users without 42 logins or ID:\n", len(AllUsers)))
	RemoveUserWithoutLogins()
	if len(AllUsers) == 0 {
		Log("⚠️  Every users got removed. Stopping script\n")
		PrintLoading()
		return
	}

	Log(fmt.Sprintf("┏ [Step 3] 🎓 [%d users remaining] Removing users that are not apprentices:\n", len(AllUsers)))
	RemoveNonApprenticeUsers()
	if len(AllUsers) == 0 {
		Log("⚠️  Every users got removed. Stopping script.\n")
		PrintLoading()
		return
	}

	Log(fmt.Sprintf("┏ [Final Step] ✉️ [%d users] Posting attendances:\n", len(AllUsers)))
	PostApprenticesAttendances()
	PrintLoading()
	Log("Stopping Watchdog")
}
