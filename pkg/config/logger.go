package config

import (
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func SetupLogger() {
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)
}
