package gateways

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	cassette "github.com/RainrainWu/quiz-cassette"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	cassetteCommandName      string = "cassette"
	cassetteEmbedColorPublic int    = 0x00ff00
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
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "cassete-description",
							Description: "description for the new cassette",
							Required:    false,
						},
					},
				},
				{
					Name:        "list",
					Description: "list your cassettes",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "show",
					Description: "show exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "cassete-id",
							Description: "the id of target cassette",
							Required:    true,
						},
					},
				},
				{
					Name:        "update",
					Description: "update exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "cassete-id",
							Description: "the id of target cassette",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "cassete-name",
							Description: "name for the new cassette",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "cassete-description",
							Description: "description for the new cassette",
							Required:    false,
						},
					},
				},
				{
					Name:        "delete",
					Description: "delete exist cassette",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "cassete-id",
							Description: "the id of target cassette",
							Required:    true,
						},
					},
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

func getDiscordUserID(i *discordgo.InteractionCreate) string {

	userID := ""
	if i.Member != nil {
		userID = i.Member.User.ID
	} else {
		userID = i.User.ID
	}
	return userID
}

func createCassetteEmbed(cst cassette.Cassette, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       cst.Name,
		Description: cst.Description,
		Type:        discordgo.EmbedTypeRich,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID",
				Value:  strconv.FormatUint(uint64(cst.ID), 10),
				Inline: true,
			},
		},
	}
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

	content, embeds := "", []*discordgo.MessageEmbed{}
	switch i.Data.Options[0].Name {
	case "create":
		_content, _embeds := handleCassetteCreateCommand(s, i)
		content, embeds = _content, append(embeds, _embeds...)
	case "list":
		embeds = append(embeds, handleCassetteListCommand(s, i)...)
	case "show":
		embeds = append(embeds, handleCassetteShowCommand(s, i)...)
	case "update":
		_content, _embeds := handleCassetteUpdateCommand(s, i)
		content, embeds = _content, append(embeds, _embeds...)
	default:
		content = fmt.Sprintf("unknown subcommand %s", i.Data.Options[0].Name)
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: content,
			Embeds:  embeds,
		},
	})
}

func handleCassetteCreateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (string, []*discordgo.MessageEmbed) {
	name := i.Data.Options[0].Options[0].StringValue()
	desc := ""
	if len(i.Data.Options[0].Options) > 1 {
		desc = i.Data.Options[0].Options[1].StringValue()
	}
	userID := getDiscordUserID(i)
	cst, _ := cassette.DBConn.NewCassette(name, desc, userID, "")
	content := fmt.Sprintf("cassete %s created", name)
	embeds := []*discordgo.MessageEmbed{
		createCassetteEmbed(cst, cassetteEmbedColorPublic),
	}
	return content, embeds
}

func handleCassetteListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) []*discordgo.MessageEmbed {

	userID := getDiscordUserID(i)
	embeds := []*discordgo.MessageEmbed{}
	csts, _ := cassette.DBConn.GetCassettesByDiscordOwner(userID)
	for _, cst := range csts {
		embeds = append(embeds, createCassetteEmbed(cst, cassetteEmbedColorPublic))
	}
	return embeds
}

func handleCassetteShowCommand(s *discordgo.Session, i *discordgo.InteractionCreate) []*discordgo.MessageEmbed {

	id := i.Data.Options[0].Options[0].UintValue()
	cst, _ := cassette.DBConn.GetCassette(uint(id))
	embeds := []*discordgo.MessageEmbed{
		createCassetteEmbed(cst, cassetteEmbedColorPublic),
	}
	return embeds
}

func handleCassetteUpdateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (string, []*discordgo.MessageEmbed) {

	id := i.Data.Options[0].Options[0].UintValue()
	mutations := map[string]interface{}{}
	if len(i.Data.Options[0].Options) > 1 {
		mutations["Name"] = i.Data.Options[0].Options[1].StringValue()
	}
	if len(i.Data.Options[0].Options) > 2 {
		mutations["Description"] = i.Data.Options[0].Options[2].StringValue()
	}
	cst, _ := cassette.DBConn.UpdateCassette(uint(id), mutations)
	content := fmt.Sprintf("cassette %s updated", cst.Name)
	embeds := []*discordgo.MessageEmbed{
		createCassetteEmbed(cst, cassetteEmbedColorPublic),
	}
	return content, embeds
}

func handleCassetteDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) string {

	id := i.Data.Options[0].Options[0].UintValue()
	cst, _ := cassette.DBConn.GetCassette(uint(id))
	cassette.DBConn.DeleteCassette(uint(id))
	return fmt.Sprintf("cassette %s updated", cst.Name)
}

func (g *discordGateway) Start() {
	discordSession, _ := discordgo.New("Bot " + g.authToken)
	g.session = discordSession
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
