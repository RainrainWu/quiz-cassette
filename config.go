package quiz_cassette

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Config ConfigSet = NewConfigSet()
)

type ConfigSet interface {
	GetDiscordAuthToken() string
	GetPostgresHost() string
	GetPostgresPort() string
	GetPostgresUser() string
	GetPostgresPassword() string
	GetPostgresDBName() string
}

type configSet struct {
	discordAuthToken string
	postgresHost     string
	postgresPort     string
	postgresUser     string
	postgresPassword string
	postgresDBName   string
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
		postgresHost:     os.Getenv("POSTGRES_HOST"),
		postgresPort:     os.Getenv("POSTGRES_PORT"),
		postgresUser:     os.Getenv("POSTGRES_USER"),
		postgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		postgresDBName:   os.Getenv("POSTGRES_DB"),
	}
	return instance
}

func (c *configSet) GetDiscordAuthToken() string {
	return c.discordAuthToken
}

func (c *configSet) GetPostgresHost() string {
	return c.postgresHost
}

func (c *configSet) GetPostgresPort() string {
	return c.postgresPort
}

func (c *configSet) GetPostgresUser() string {
	return c.postgresUser
}

func (c *configSet) GetPostgresPassword() string {
	return c.postgresPassword
}

func (c *configSet) GetPostgresDBName() string {
	return c.postgresDBName
}
