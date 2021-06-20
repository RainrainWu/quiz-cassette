package gateways

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cassette "github.com/RainrainWu/quiz-cassette"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	cassetteCommandName string = "cassette"
)

var (
	Discord  DiscordGateway = newDiscordSession()
	commands                = []*discordgo.ApplicationCommand{
		{
			Name:        cassetteCommandName,
			Description: "commands for using cassette",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "create",
					Description: "create new cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "cassete-name",
							Description: "name for the new cassette",
							Required:    true,
						},
					},
				},
				{
					Name:        "update",
					Description: "update exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "show",
					Description: "show exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "delete",
					Description: "delete exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
	slashCommandHandlerMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		cassetteCommandName: handleCassetteCommand,
	}
)

type DiscordGateway interface {
	createSlashCommand()
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

func (g *discordGateway) createSlashCommand() {

	for _, cmd := range commands {
		_, err := g.session.ApplicationCommandCreate(cassette.Config.GetDiscordAppID(), "", cmd)
		if err != nil {
			cassette.Logger.Fatal(
				"discord slash command create failed",
				zap.String("name", cmd.Name),
				zap.String("err", err.Error()),
			)
		}
	}
}

func handleEcho(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	s.ChannelMessageSend(m.ChannelID, m.Content)
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	cassette.Logger.Info("recieve " + i.Data.Name)
	if handler, ok := slashCommandHandlerMap[i.Data.Name]; ok {
		handler(s, i)
	} else {
		cassette.Logger.Warn(
			"undefined slash commands",
			zap.String("command", i.Data.Name),
		)
	}
}

func handleCassetteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	content := ""
	switch i.Data.Options[0].Name {
	case "create":
		name := i.Data.Options[0].Options[0].StringValue()
		content = fmt.Sprintf("cassete %s created", name)
	default:
		content = fmt.Sprintf("unknown subcommand %s", i.Data.Options[0].Name)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: content,
		},
	})
}

func (g *discordGateway) Start() {
	discordSession, _ := discordgo.New("Bot " + g.authToken)
	g.session = discordSession
	g.session.AddHandler(handleEcho)
	g.session.AddHandler(handleSlashCommand)
	g.session.Identify.Intents = discordgo.IntentsGuildMessages
	g.createSlashCommand()

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
