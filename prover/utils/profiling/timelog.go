package profiling

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func LogTimer(msg string, args ...any) func() {
	msgString := fmt.Sprintf(msg, args...)
	logrus.Infoln(msgString + " - START")
	timer := time.Now()
	return func() {
		logrus.Infof(msgString+" - DONE - took %v", time.Since(timer))
	}
}
