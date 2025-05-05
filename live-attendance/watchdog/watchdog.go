package watchdog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"watchdog/apiManager"
	"watchdog/config"
)

type UserV2 struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

var acceptEvents bool = false
var acceptEventsMutex sync.Mutex

func FetchMissingFields(login string, userID string) (string, string) {
	resp, err := apiManager.GetClient(config.FTv2).Get(fmt.Sprintf("/users?filter[id]=%s&filter[login]=%s", userID, strings.ToLower(login)))
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		return login, userID
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Log(fmt.Sprintf("ERROR: couldn't fetch user: %s\n", resp.Status))
		return login, userID
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s\n", err.Error()))
		return login, userID
	}

	var res []UserV2
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("ERROR: %s", err.Error()))
		return login, userID
	}

	if len(res) == 0 {
		Log(fmt.Sprintf("ERROR: user (%s|%s) not found", login, userID))
		return login, userID
	}

	if len(res) > 1 {
		Log(fmt.Sprintf("ERROR: many user found with (%s|%s)", login, userID))
		return login, userID
	}

	return res[0].Login, strconv.FormatInt(int64(res[0].ID), 10)
}

func GetAllowEvents() bool {
	dest := false
	acceptEventsMutex.Lock()
	dest = acceptEvents
	acceptEventsMutex.Unlock()
	return dest
}

func AllowEvents(isAllowed bool) {
	if isAllowed {
		Log("[WATCHDOG] 🟢 accepting incomming events")
	} else {
		Log("[WATCHDOG] 🔴 refusing incomming events")
	}
	acceptEventsMutex.Lock()
	acceptEvents = isAllowed
	acceptEventsMutex.Unlock()
}

func CreateNewUser(userID int, accessControlUsername string) (User, error) {
	user := User{
		ControlAccessID:   userID,
		ControlAccessName: accessControlUsername,
	}

	resp, err := apiManager.GetClient(config.AccessControl).Get(fmt.Sprintf("/users/%d", userID))
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s", err.Error()))
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s", err.Error()))
		os.Exit(1)
	}

	var res UserResponse
	err = json.Unmarshal(bodyBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s", err.Error()))
		os.Exit(1)
	}
	user.Login42 = res.Properties.Login
	user.ID42 = res.Properties.ID
	if user.Login42 == "" && user.ID42 == "" {
		return User{}, fmt.Errorf("user (%s) has no Login42 AND no ID42", accessControlUsername)
	}

	if user.Login42 == "" || user.ID42 == "" {
		user.Login42, user.ID42 = FetchMissingFields(user.Login42, user.ID42)
	}

	if user.Login42 == "" || user.ID42 == "" {
		return User{}, fmt.Errorf("failed to fetch Login42('%s') OR ID42('%s')", user.Login42, user.ID42)
	}

	user.IsApprentice = false
	for _, projectID := range config.ConfigData.ApiV2.ApprenticeProjects {
		if isProjectOngoing(user.Login42, projectID) {
			user.IsApprentice = true
			break
		}
	}
	if user.IsApprentice {
		Log(fmt.Sprintf("[WATCHDOG] 📋 Created a new user: %s is an apprentice", user.Login42))
	} else {
		Log(fmt.Sprintf("[WATCHDOG] 📋 Created a new user: %s is a basic student", user.Login42))
	}
	return user, nil
}

func UpdateUserAccess(userID int, accessControlUsername string, timeStamp time.Time, doorName string) {
	var err error
	AllUsersMutex.Lock()
	user, exist := AllUsers[userID]
	if !exist {
		user, err = CreateNewUser(userID, accessControlUsername)
		if err != nil {
			Log(fmt.Sprintf("[WATCHDOG] ❌ Failed to create user: %s\n", err.Error()))
			AllUsersMutex.Unlock()
			return
		}
	}
	AllUsers[userID] = user
	AllUsersMutex.Unlock()

	if !acceptEvents {
		Log(fmt.Sprintf("[WATCHDOG] 🚪 User %s used door %s, but watchdog is sleeping", user.Login42, doorName))
		return
	}
	Log(fmt.Sprintf("[WATCHDOG] 🚪 User %s used door %s", user.Login42, doorName))
	if user.FirstAccess.IsZero() || user.FirstAccess.After(timeStamp) {
		user.FirstAccess = timeStamp
		user.Duration = user.LastAccess.Sub(user.FirstAccess)
	}
	if user.LastAccess.IsZero() || user.LastAccess.Before(timeStamp) {
		user.LastAccess = timeStamp
		user.Duration = user.LastAccess.Sub(user.FirstAccess)
	}
	AllUsersMutex.Lock()
	AllUsers[userID] = user
	AllUsersMutex.Unlock()
}

func PrintUsersTimers() {
	parisLoc, _ := time.LoadLocation("Europe/Paris")
	AllUsersMutex.Lock()
	total := len(AllUsers)
	i := 1
	if total == 0 {
		Log("[WATCHDOG] No users saved")
	} else {
		Log("[WATCHDOG] ┌─ Users status:")
	}
	for _, value := range AllUsers {
		box := getBoxChar(i, total)
		if value.FirstAccess.IsZero() {
			Log(fmt.Sprintf("[WATCHDOG] %s %8s: %s ┆ Total : %s\n",
				box,
				value.Login42,
				"  No badge usage yet ",
				formatDuration(value.Duration),
			))
		} else {
			Log(fmt.Sprintf("[WATCHDOG] %s %8s: %s -> %s ┆ Total : %s\n",
				box,
				value.Login42,
				value.FirstAccess.In(parisLoc).Format("15h04m05s"),
				value.LastAccess.In(parisLoc).Format("15h04m05s"),
				formatDuration(value.Duration),
			))
		}
		i++
	}
	AllUsersMutex.Unlock()
}

func isProjectOngoing(login string, projectID string) bool {
	resp, err := apiManager.GetClient(config.FTv2).Get(fmt.Sprintf("/users/%s/projects/%s/teams?sort=-created_at", login, projectID))
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s\n", err.Error()))
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s\n", err.Error()))
		return false
	}

	var res []ProjectResponse
	err = json.Unmarshal(respBytes, &res)
	if err != nil {
		Log(fmt.Sprintf("[WATCHDOG] ERROR: %s", err.Error()))
		os.Exit(1)
	}
	return len(res) >= 1 && res[0].Status == "in_progress"
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
		return "└──"
	} else {
		return "├──"
	}
}

func PostApprenticesAttendances() {
	msg := ""
	AllUsersMutex.Lock()
	total := len(AllUsers)
	i := 1
	if total == 0 {
		Log("[WATCHDOG] Posting Attendances: no users registered")
		return
	}
	Log("[WATCHDOG] ┌─ Posting Attendances:")
	for _, value := range AllUsers {
		var resp *http.Response
		box := getBoxChar(i, total)
		if value.FirstAccess.IsZero() {
			i++
			continue
		}
		id42, err := strconv.ParseInt(value.ID42, 10, 64)
		if err != nil {
			msg = fmt.Sprintf("couldn't convert string to int \"%s\"", value.ID42)
		} else {
			_ = id42
			if value.FirstAccess == value.LastAccess {
				msg = "(Used badge only once)"
				err = fmt.Errorf("badge error")
			} else if value.Duration < time.Duration(60*10) {
				msg = fmt.Sprintf("(too short duration %.2f minutes)", value.Duration.Minutes())
				err = fmt.Errorf("badge error")
			} else if !value.IsApprentice {
				msg = "user is not an apprentice"
				err = fmt.Errorf("user is not an apprentice")
			} else if config.ConfigData.Attendance42.AutoPost {
				resp, err = apiManager.GetClient(config.FTAttendance).Post("/attendances", APIAttendance{
					Begin_at:  value.FirstAccess.UTC().Format(time.RFC3339),
					End_at:    value.LastAccess.UTC().Format(time.RFC3339),
					Source:    "access-control",
					Campus_id: 41,
					User_id:   int(id42),
				})
				if err != nil {
					msg = err.Error()
				} else if resp.StatusCode != http.StatusOK {
					msg = resp.Status
					err = fmt.Errorf("invalid status code")
				} else {
					msg = resp.Status
				}
			} else {
				msg = "AUTOPOST is OFF"
				err = fmt.Errorf("autopost is off")
			}
		}
		if err != nil {
			Log(fmt.Sprintf("[WATCHDOG] %s ❌ Posted attendance for %-8s (%s): %s\n", box, value.Login42, formatDuration(value.Duration), msg))
		} else {
			Log(fmt.Sprintf("[WATCHDOG] %s ✅ Posted attendance for %-8s (%s): %s\n", box, value.Login42, formatDuration(value.Duration), msg))
		}
		value.FirstAccess = time.Time{}
		value.LastAccess = time.Time{}
		value.Duration = value.FirstAccess.Sub(value.LastAccess)
		AllUsers[value.ControlAccessID] = value
		i++
	}
	AllUsersMutex.Unlock()
}
