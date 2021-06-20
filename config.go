package quizdeck

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Config ConfigSet = NewConfigSet()
)

type ConfigSet interface {
	GetDiscordAuthToken() string
}

type configSet struct {
	discordAuthToken string
}

func NewConfigSet() ConfigSet {
	err := godotenv.Load()
	if err != nil {
		Logger.Warn(
			"error loading env variables",
		)
	}
	instance := &configSet{
		discordAuthToken: os.Getenv("DISCORD_AUTH_TOKEN"),
	}
	return instance
}

func (c *configSet) GetDiscordAuthToken() string {
	return c.discordAuthToken
}
