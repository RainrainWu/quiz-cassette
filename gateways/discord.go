package gateways

import (
	"os"
	"os/signal"
	"syscall"

	cassette "github.com/RainrainWu/quiz-cassette"
	"github.com/bwmarrin/discordgo"
)

var (
	Discord DiscordGateway = newDiscordSession()
)

type DiscordGateway interface {
	Start()
}

type discordGateway struct {
	session   *discordgo.Session
	authToken string
}

func newDiscordSession() DiscordGateway {

	instance := &discordGateway{
		session:   nil,
		authToken: cassette.Config.GetDiscordAuthToken(),
	}
	return instance
}

func handleEcho(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	s.ChannelMessageSend(m.ChannelID, m.Content)
}

func (g *discordGateway) Start() {
	discordSession, _ := discordgo.New("Bot " + g.authToken)
	g.session = discordSession
	g.session.AddHandler(handleEcho)
	g.session.Identify.Intents = discordgo.IntentsGuildMessages

	err := g.session.Open()
	if err != nil {
		cassette.Logger.Warn("error opening connection, " + err.Error())
		return
	}

	cassette.Logger.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	g.session.Close()
}
