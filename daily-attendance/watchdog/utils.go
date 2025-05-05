package watchdog

import (
	"net/url"
	"time"
)

func formatTimeForURL(t time.Time) string {
	return url.QueryEscape(t.Format("2006-01-02 15:04"))
}

func getBoxChar(i int, len int) string {
	if i == len {
		return "┗"
	} else {
		return "┠"
	}
}
