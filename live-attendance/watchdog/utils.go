package watchdog

import (
	"fmt"
	"time"
)

func formatDuration(d time.Duration) string {
	// Optional: Handle negative durations if they can occur
	// if d < 0 {
	//     d = -d // or return an error/special string
	// }

	// Round to nearest second for cleaner output, optional
	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		// Includes hours: hh:mm:ss
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	} else if m > 0 {
		// Includes minutes: mm:ss
		return fmt.Sprintf("%dm%02ds", m, s)
	} else {
		// Seconds only: ss
		return fmt.Sprintf("%ds", s)
	}
}
