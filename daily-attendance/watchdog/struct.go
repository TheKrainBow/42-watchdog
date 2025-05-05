package watchdog

import (
	"time"
)

var StepNumber = 0

var totalSteps = 0
var actualStep = 0
var durationTotalSteps = -1
var durationActualSteps = 0

var loginTotalSteps = -1
var loginActualSteps = 0

var nonApprenticeTotalSteps = -1
var nonApprenticeActualSteps = 0

var postAttendanceTotalSteps = -1
var postAttendanceActualSteps = 0

type User struct {
	ControlAccessID   int
	ControlAccessName string
	Login42           string
	ID42              string
	IsApprentice      bool
	FirstAccess       time.Time
	LastAccess        time.Time
	Duration          time.Duration
}

type ProjectResponse struct {
	Status string `json:"status"`
}

var AllUsers map[int]User

type UserAccess struct {
	UserID      int       `json:"id"`
	FirstAccess time.Time `json:"first_access"`
	LastAccess  time.Time `json:"last_access"`
}

type Property42 struct {
	Login string `json:"ft_login"`
	ID    string `json:"ft_id"`
}

type UserResponse struct {
	Properties Property42 `json:"properties"`
}

type EventResponse struct {
	Data []Event `json:"data"`
}

type Event struct {
	User     *int      `json:"user"`      // pointer since it can be null
	DateTime string    `json:"date_time"` // or time.Time if you want to parse it
	Data     EventData `json:"data"`
}

type EventData struct {
	DoorName    string `json:"door_name"`
	UserName    string `json:"user_name"`
	DeviceName  string `json:"device_name"`
	BadgeNumber string `json:"badge_number"`
}
