package main

import (
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

var Recipients = []string{
	"heinz@42nice.fr",
	// "tac@42nice.fr",
}

const (
	NO_BADGE       string = "User didn't badged yet"
	BADGED_ONCE    string = "User badge only once"
	NOT_APPRENTICE string = "User is not an apprentice"

	APPRENTICE_NO_BADGE    string = "Apprentice didn't badged yet"
	APPRENTICE_BADGED_ONCE string = "Apprentice badged only once"
	POSTED                 string = "Posted"
	POST_ERROR             string = "Post returned an error"
)

type User struct {
	ControlAccessID   int           `json:"control_access_id"`
	ControlAccessName string        `json:"control_access_name"`
	Login42           string        `json:"login_42"`
	ID42              string        `json:"id_42"`
	IsApprentice      bool          `json:"is_apprentice"`
	FirstAccess       time.Time     `json:"first_access"`
	LastAccess        time.Time     `json:"last_access"`
	Duration          time.Duration `json:"duration"`
	Error             error
	Status            string
}

func addLogToMail(htmlBody *strings.Builder, user User, loc *time.Location) {
	color := "green"
	firstColor := "green"
	lastColor := "green"
	durationColor := "green"
	emoji := "✅"

	msg := user.Status
	if user.Error != nil {
		msg = user.Error.Error()
	}

	first := user.FirstAccess.In(loc)
	if first.Before(time.Date(first.Year(), first.Month(), first.Day(), 8, 0, 0, 0, loc)) {
		firstColor = "orange"
	}

	last := user.LastAccess.In(loc)
	if last.After(time.Date(last.Year(), last.Month(), last.Day(), 20, 0, 0, 0, loc)) {
		lastColor = "orange"
	}

	if user.Duration < 5*time.Hour {
		durationColor = "red"
	} else if user.Duration < 7*time.Hour {
		durationColor = "orange"
	}

	if user.Status != POSTED {
		color = "red"
		firstColor = "red"
		lastColor = "red"
		emoji = "❌"
		durationColor = "red"
	}

	htmlBody.WriteString(`<tr><td style="white-space: pre; font-family: Menlo, Consolas, 'Courier New', monospace; font-size: 13px; padding: 1px 20px; line-height: 1;">`)
	htmlBody.WriteString(`<span style="color: green;">` + emoji + `</span> `)
	htmlBody.WriteString(`<span style="color:` + color + `;">` + fmt.Sprintf("%-8s", user.Login42) + `</span>: `)
	htmlBody.WriteString(`<span style="color:` + firstColor + `;">` + first.Format("15:04:05") + `</span>-`)
	htmlBody.WriteString(`<span style="color:` + lastColor + `;">` + last.Format("15:04:05") + `</span> `)
	htmlBody.WriteString(`<span style="color:` + durationColor + `;">` + formatDuration(user.Duration) + `</span> — `)
	htmlBody.WriteString(`<span style="color:` + color + `;">` + msg + `</span>`)
	htmlBody.WriteString(`</td></tr>`)
}

func formatDuration(d time.Duration) string {
	// Round to nearest second for cleaner output, optional
	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 && h < 10 {
		return fmt.Sprintf(" (%dh%02dm%02ds) ", h, m, s)
	} else if h >= 10 {
		return fmt.Sprintf("(%02dh%02dm%02ds) ", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("  (%dm%02ds)   ", m, s)
	} else if s < 10 {
		return fmt.Sprintf("    (%ds)    ", s)
	} else {
		return fmt.Sprintf("    (%02ds)    ", s)
	}
}

func PostApprenticesAttendances() {
	parisLoc, _ := time.LoadLocation("Europe/Paris")
	sortedUser := map[string][]User{
		NO_BADGE: {
			{Login42: "abianchi", FirstAccess: time.Time{}, LastAccess: time.Time{}, Duration: 0, Status: NO_BADGE},
			{Login42: "cjones", FirstAccess: time.Time{}, LastAccess: time.Time{}, Duration: 0, Status: NO_BADGE},
		},
		BADGED_ONCE: {
			{Login42: "dsmith", FirstAccess: time.Date(2025, 7, 2, 9, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 9, 0, 0, 0, time.UTC), Duration: 0, Status: BADGED_ONCE},
		},
		NOT_APPRENTICE: {
			{Login42: "emartin", FirstAccess: time.Date(2025, 7, 2, 10, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 18, 0, 0, 0, time.UTC), Duration: 8 * time.Hour, Status: NOT_APPRENTICE},
		},
		APPRENTICE_NO_BADGE: {
			{Login42: "fgomez", IsApprentice: true, FirstAccess: time.Time{}, LastAccess: time.Time{}, Duration: 0, Status: APPRENTICE_NO_BADGE},
		},
		APPRENTICE_BADGED_ONCE: {
			{Login42: "hlee", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 8, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 8, 0, 0, 0, time.UTC), Duration: 0, Status: APPRENTICE_BADGED_ONCE},
		},
		POSTED: {
			{Login42: "jdoe", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 8, 30, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 17, 0, 0, 0, time.UTC), Duration: 10*time.Hour + 30*time.Minute, Status: POSTED},
			{Login42: "kchan", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 9, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 16, 45, 0, 0, time.UTC), Duration: 7*time.Hour + 45*time.Minute, Status: POSTED},
			{Login42: "maagosti", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 9, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 16, 45, 0, 0, time.UTC), Duration: 4*time.Hour + 45*time.Minute, Status: POSTED},
			{Login42: "ltcherep", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 9, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 20, 07, 0, 0, time.UTC), Duration: 5*time.Hour + 45*time.Minute, Status: POSTED},
			{Login42: "gkzodis", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 7, 48, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 20, 07, 0, 0, time.UTC), Duration: 4*time.Hour + 45*time.Minute, Status: POSTED},
		},
		POST_ERROR: {
			{Login42: "lpetit", IsApprentice: true, FirstAccess: time.Date(2025, 7, 2, 10, 0, 0, 0, time.UTC), LastAccess: time.Date(2025, 7, 2, 14, 15, 0, 0, time.UTC), Duration: 4*time.Hour + 15*time.Minute, Status: POST_ERROR, Error: fmt.Errorf("Status not OK")},
		},
	}

	for status, users := range sortedUser {
		sort.Slice(users, func(i, j int) bool {
			return users[i].Login42 < users[j].Login42
		})
		sortedUser[status] = users
	}

	// LOG AND MAIL APPRENTICES:

	var htmlBody strings.Builder
	today := time.Now()
	htmlBody.WriteString("<h2>Watchdog Daily Report – " + today.Format("02/01/2006") + "</h2>")
	htmlBody.WriteString(`
		<table style="border:2px solid #ccc; padding: 8px; border-collapse:collapse; background:#f9f9f9;">
	`)
	htmlBody.WriteString(`<tr><td style="white-space: pre; font-size: 13px; padding: 1px; padding-left: 20px; padding-right: 20px; line-height: 1;">  </td></tr>`)
	if len(sortedUser[POSTED]) > 0 {
		for _, user := range sortedUser[POSTED] {
			addLogToMail(&htmlBody, user, parisLoc)
		}
		htmlBody.WriteString(`<tr><td style="white-space: pre; font-size: 13px; padding: 1px; padding-left: 20px; padding-right: 20px; line-height: 1;">  </td></tr>`)
	}

	if len(sortedUser[POST_ERROR]) > 0 {
		for _, user := range sortedUser[POST_ERROR] {
			addLogToMail(&htmlBody, user, parisLoc)
		}
	}

	if len(sortedUser[APPRENTICE_BADGED_ONCE]) > 0 {
		for _, user := range sortedUser[APPRENTICE_BADGED_ONCE] {
			addLogToMail(&htmlBody, user, parisLoc)
		}
	}

	if len(sortedUser[APPRENTICE_NO_BADGE]) > 0 {
		for _, user := range sortedUser[APPRENTICE_NO_BADGE] {
			addLogToMail(&htmlBody, user, parisLoc)
		}
	}
	htmlBody.WriteString(`<tr><td style="white-space: pre; font-size: 13px; padding: 1px; padding-left: 20px; padding-right: 20px; line-height: 1;">  </td></tr>`)
	htmlBody.WriteString(`</table><p style="font-size:11px; color:#888;">Generated by Watchdog at ` + today.Format("15:04:05") + ` - Timezone is CEST</p>`)

	fmt.Println("🚀 Attempting to send:")
	fmt.Println(htmlBody.String())
	Send(Recipients, fmt.Sprintf("Watchdog – Daily Report - %s", time.Now().Format("02/01/2006")), htmlBody.String(), true)
	fmt.Println("✅ Success!") // this won't print if it hangs
}

func main() {
	// var htmlBody strings.Builder
	fmt.Println("📤 Creating dialer")
	d := gomail.NewDialer("smtp-relay.gmail.com", 587, "", "")
	d.TLSConfig = &tls.Config{ServerName: "smtp-relay.gmail.com"}

	fmt.Println("📨 Preparing message")
	Init(ConfMailer{
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

	PostApprenticesAttendances()
}
