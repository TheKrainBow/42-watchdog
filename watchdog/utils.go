package watchdog

import (
	"net/url"
	"time"
)

func formatTimeForURL(t time.Time) string {
	return url.QueryEscape(t.Format("2006-01-02 15:04"))
}
