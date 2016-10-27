package log

import (
	"github.com/catpie/logrus"
)

var (
	log = logrus.New()
)

func Init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}
