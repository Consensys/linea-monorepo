package signal

import (
	"os"
	"os/signal"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Registers a signal handler that dumps the StackTraces upon receiving a
// SIGUSR1.
func RegisterStackTraceDumpHandler(v os.Signal) {

	sigChan := make(chan os.Signal, 1<<8)

	go func() {
		for range sigChan {
			logrus.Infof("received signal %v -> dumping the stack traces", v)
			// Allocate a buffer to hold the stack data
			buf := make([]byte, 1024*1024)
			n := runtime.Stack(buf, true) // 'true' gets all goroutines
			logrus.Infof("stack trace: %s", string(buf[:n]))
		}
	}()

	signal.Notify(sigChan, v)
	logrus.Infof("Registered the signal handler for %s. If you send that signal to the process it will dump the stack traces", v)
}
