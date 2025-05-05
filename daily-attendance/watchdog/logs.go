package watchdog

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var logFile *os.File

var isFirstPrint = true

func loadingBar(title string, index, total int, onScreenSize int, withStep bool, withNewline bool) {
	if index > total {
		index = total
	}
	percent := float64(index) / float64(total)
	fmt.Printf("%s[\033[1;32m", title)
	for i := 0; i <= (onScreenSize - 2); i++ {
		if total != -1 && float64(i)/float64(onScreenSize-2) <= percent {
			fmt.Printf("■")
		} else {
			fmt.Printf(" ")
		}
	}
	fmt.Printf("\033[0m]")
	if withStep {
		if total <= 0 {
			fmt.Printf(" [  ...  ]")
		} else {
			fmt.Printf(" [%3d/%-3d]", index, total)
		}
	}
	if total == -1 {
		fmt.Printf("\033[90m Waiting")
	} else if total == index && total != -1 {
		fmt.Printf("\033[1;32m Done")
	} else {
		fmt.Printf("\033[33m On-going")
	}
	if withNewline {
		fmt.Printf("\n")
	}
	fmt.Printf("\033[0m")
}

func PrintLoading() {
	if isFirstPrint {
		isFirstPrint = false
		fmt.Print("\033[1A\033[2K")
	} else {
		fmt.Print("\033[1A\033[2K")
		fmt.Print("\033[1A\033[2K")
		fmt.Print("\033[1A\033[2K")
		fmt.Print("\033[1A\033[2K")
		fmt.Print("\033[1A\033[2K")
	}
	loadingBar("Total progress:  ", actualStep, totalSteps, 22, true, true)
	loadingBar("Short Duration:  ", durationActualSteps, durationTotalSteps, 22, true, true)
	loadingBar("No Login:        ", loginActualSteps, loginTotalSteps, 22, true, true)
	loadingBar("Non Apprentices: ", nonApprenticeActualSteps, nonApprenticeTotalSteps, 22, true, true)
	loadingBar("Post Attendance: ", postAttendanceActualSteps, postAttendanceTotalSteps, 22, true, true)
	actualStep++
}

func InitLogs(logPath string) (err error) {
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	return nil
}

func Log(msg string) {
	if msg == "" {
		fmt.Fprintf(logFile, "\n")
	} else {
		msg = strings.TrimRight(msg, "\n")
		fmt.Fprintf(logFile, "[%s] %s\n", time.Now().Format("06/01/02 - 15:04:05"), msg)
	}
}

func CloseLogs() {
	logFile.Close()
}
