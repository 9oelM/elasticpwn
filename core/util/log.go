package EPUtils

import (
	"log"
	"os"
	"time"
)

func timeIn(name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}

func EPLog(l *log.Logger, msg string) {
	l.SetPrefix(timeIn("Asia/Seoul").Format("2006-01-02 15:04:05") + " [EP] ")
	l.Print(msg)
}

var L = log.New(os.Stdout, "", 0)

func EPLogger(message string) {
	EPLog(L, message)
}
