package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// SetupLogger initializes the logger with the given log level and log file.
func SetupLogger(level logLevel) error {
	// Format the style of the logger
	formatter := &logrus.TextFormatter{
		TimestampFormat: "15-01-2018 15:04:05.000000",
		FullTimestamp:   true,
	}

	// TODO see https://github.com/sirupsen/logrus/issues/894#issuecomment-1284051207 for a
	// more elegant solution
	// NOTE (23/10/2023): unknown what the previous author of the code meant with the issue above
	// and what problem he/she was trying to solve.
	logrus.SetFormatter(formatter)

	logrus.SetOutput(os.Stdout)

	switch level {
	case _logTrace:
		logrus.SetLevel(logrus.TraceLevel)
	case _logDebug:
		logrus.SetLevel(logrus.DebugLevel)
	case _logInfo:
		logrus.SetLevel(logrus.InfoLevel)
	case _logWarn:
		logrus.SetLevel(logrus.WarnLevel)
	case _logError:
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		// this should never happen, as this config variable is checked when calling either
		// ParseContractGenConfig or ParseProverConfig
		return fmt.Errorf("unknown log level: %d", level)
	}

	logrus.Infof("Log level %v", level)

	return nil
}
