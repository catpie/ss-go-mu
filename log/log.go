package log

import (
	"github.com/catpie/logrus"
)

var (
	Log = logrus.New()
)

func Init() {
	Log.Formatter = new(logrus.JSONFormatter)
	Log.Formatter = new(logrus.TextFormatter) // default
	Log.Level = logrus.DebugLevel
}
