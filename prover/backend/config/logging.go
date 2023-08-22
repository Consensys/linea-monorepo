package config

import (
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

const (
	// envconfig prefix used by the app
	prefix = "LOG"
)

// Global configuration for the prover
type LoggingSpec struct {
	LogLevel string `envconfig:"LEVEL" default:"info"`
}

func InitLogger() {
	conf := &LoggingSpec{}

	// Parse the log level from the environmenent variable
	envconfig.MustProcess(prefix, conf)

	// Format the style of the logger
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "15-01-2018 15:04:05.000000"
	Formatter.FullTimestamp = true
	Formatter.ForceColors = true
	logrus.SetFormatter(Formatter)

	switch conf.LogLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		utils.Panic("Expected log-level to either of trace|debug|info|warn|error")
	}

	// Print the log level
	logrus.Infof("Log level %v", conf.LogLevel)
}

// Initializes the logger
func InitApp() {

	InitLogger()

	proConf := MustGetProver()
	ethConf := MustGetLayer2()

	logrus.Infof("Eth config : %++v", ethConf)
	logrus.Infof("Prover config : %++v", proConf)
}
