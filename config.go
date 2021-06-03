package quizdeck

import (
	"github.com/joho/godotenv"
)

var (
	Config ConfigSet = NewConfigSet()
)

type ConfigSet interface{}

type configSet struct{}

func NewConfigSet() ConfigSet {
	err := godotenv.Load()
	if err != nil {
		Logger.Warn(
			"error loading .env file, current environment " +
				"variables would be used directly",
		)
	}
	instance := &configSet{}
	return instance
}
