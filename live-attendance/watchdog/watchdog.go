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
	"watchdog/config"

	apiManager "github.com/TheKrainBow/go-api"
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
		Log("[WATCHDOG] 🟢 accepting incoming events")
	} else {
		Log("[WATCHDOG] 🔴 refusing incoming events")
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
	defer AllUsersMutex.Unlock()

	total := len(AllUsers)
	i := 1
	if total == 0 {
		Log("[WATCHDOG] No users saved")
		return
	}

	Log("[WATCHDOG] ┌─ Users status:")

	var noBadgeUsers, badgeUsers []User
	for _, user := range AllUsers {
		if user.FirstAccess.IsZero() {
			noBadgeUsers = append(noBadgeUsers, user)
		} else {
			badgeUsers = append(badgeUsers, user)
		}
	}

	for _, user := range noBadgeUsers {
		boxChar := getBoxChar(i, total)
		Log(fmt.Sprintf("[WATCHDOG] %s %8s: %s ┆ Total : %s\n",
			boxChar,
			user.Login42,
			"  No badge usage yet  ",
			formatDuration(user.Duration),
		))
		i++
	}

	if len(badgeUsers) > 0 {
		Log("[WATCHDOG] ├──────── Users that badged today")
	}

	i--
	for _, user := range badgeUsers {
		boxChar := getBoxChar(i, total)
		Log(fmt.Sprintf("[WATCHDOG] %s %8s: %s -> %s ┆ Total : %s\n",
			boxChar,
			user.Login42,
			user.FirstAccess.In(parisLoc).Format("15h04m05s"),
			user.LastAccess.In(parisLoc).Format("15h04m05s"),
			formatDuration(user.Duration),
		))
		i++
	}
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

func formatPostInfo(logs map[string][]string, errorType string, user User, loc *time.Location, emoji string, msg string) map[string][]string {
	apprenticeEmoji := "👤"
	if user.IsApprentice {
		apprenticeEmoji = "🎓"
	}
	logs[apprenticeEmoji+errorType] = append(logs[apprenticeEmoji+errorType], fmt.Sprintf(
		"[WATCHDOG] [POST] ├── %s %-8s: %s-%s (%s) — %s\n",
		emoji,
		user.Login42,
		user.FirstAccess.In(loc).Format("15:04:05"),
		user.LastAccess.In(loc).Format("15:04:05"),
		formatDuration(user.Duration),
		msg,
	))
	return logs
}

const NO_BADGE string = "No badge"
const BADGED_ONCE string = "Badged once"
const NOT_APPRENTICE string = "Not apprentice"
const POSTED string = "posted"
const OTHER string = "Other"

func resetUserDuration(user User) {
	user.FirstAccess = time.Time{}
	user.LastAccess = time.Time{}
	user.Duration = 0
	AllUsers[user.ControlAccessID] = user
}

func PostApprenticesAttendances() {
	parisLoc, _ := time.LoadLocation("Europe/Paris")
	logs := map[string][]string{}
	AllUsersMutex.Lock()
	defer AllUsersMutex.Unlock()
	total := len(AllUsers)
	i := 0
	if total == 0 {
		Log("[WATCHDOG] [POST] Posting Attendances: no users registered")
		return
	}
	Log("[WATCHDOG] [POST] ┌─ Posting Attendances:")
	for _, user := range AllUsers {
		i++
		if user.FirstAccess.IsZero() {
			logs = formatPostInfo(logs, NO_BADGE, user, parisLoc, "❌", "No badge usage yet")
			resetUserDuration(user)
			continue
		}

		id42, _ := strconv.ParseInt(user.ID42, 10, 64)

		if user.FirstAccess == user.LastAccess {
			logs = formatPostInfo(logs, BADGED_ONCE, user, parisLoc, "❌", "Used badge only once")
			resetUserDuration(user)
			continue
		}

		// if user.Duration < time.Duration(60*10) {
		// 	logs = formatPostInfo(logs, OTHER, user, parisLoc, "❌", fmt.Sprintf("(too short duration %.2f minutes)", user.Duration.Minutes()))
		// 	resetUserDuration(user)
		// 	continue
		// }

		if !user.IsApprentice {
			logs = formatPostInfo(logs, NOT_APPRENTICE, user, parisLoc, "❌", "user is not an apprentice")
			resetUserDuration(user)
			continue
		}

		if !config.ConfigData.Attendance42.AutoPost {
			logs = formatPostInfo(logs, POSTED, user, parisLoc, "❌", "AUTOPOST is OFF")
			resetUserDuration(user)
			continue
		}

		resp, err := apiManager.GetClient(config.FTAttendance).Post("/attendances", APIAttendance{
			Begin_at:  user.FirstAccess.UTC().Format(time.RFC3339),
			End_at:    user.LastAccess.UTC().Format(time.RFC3339),
			Source:    "access-control",
			Campus_id: 41,
			User_id:   int(id42),
		})

		if err != nil {
			logs = formatPostInfo(logs, POSTED, user, parisLoc, "❌", err.Error())
			resetUserDuration(user)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			logs = formatPostInfo(logs, POSTED, user, parisLoc, "❌", resp.Status)
			resetUserDuration(user)
			continue
		}

		logs = formatPostInfo(logs, POSTED, user, parisLoc, "✅", resp.Status)
		resetUserDuration(user)
	}

	if len(logs["👤"+NO_BADGE]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Students: No badge used today")
		for _, line := range logs["👤"+NO_BADGE] {
			Log(line)
		}
	}

	if len(logs["👤"+BADGED_ONCE]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Students: Used badge only once")
		for _, line := range logs["👤"+BADGED_ONCE] {
			Log(line)
		}
	}

	if len(logs["👤"+NOT_APPRENTICE]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Students: Not an apprentice")
		for _, line := range logs["👤"+NOT_APPRENTICE] {
			Log(line)
		}
	}

	if len(logs["👤"+OTHER]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Students: Other issues")
		for _, line := range logs["👤"+OTHER] {
			Log(line)
		}
	}

	if len(logs["🎓"+NO_BADGE]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Apprentices: No badge used today")
		for _, line := range logs["🎓"+NO_BADGE] {
			Log(line)
		}
	}

	if len(logs["🎓"+BADGED_ONCE]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Apprentices: Used badge only once")
		for _, line := range logs["🎓"+BADGED_ONCE] {
			Log(line)
		}
	}

	if len(logs["🎓"+POSTED]) > 0 {
		Log("[WATCHDOG] [POST] ├──────── Apprentices: Posts")
		for _, line := range logs["🎓"+POSTED] {
			Log(line)
		}
	}
	Log("[WATCHDOG] [POST] └── Done")
}
