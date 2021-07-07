package gateways

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/RainrainWu/quizdeck"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	deckCommandName      string = "deck"
	deckEmbedColorPublic int    = 0x00ff00
)

var (
	Discord  DiscordGateway = newDiscordSession()
	commands                = []*discordgo.ApplicationCommand{
		{
			Name:        deckCommandName,
			Description: "commands for using deck",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "create",
					Description: "create new deck",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "deck-name",
							Description: "name for the new deck",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "deck-description",
							Description: "description for the new deck",
							Required:    false,
						},
					},
				},
				{
					Name:        "list",
					Description: "list your decks",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "show",
					Description: "show exist deck",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "deck-id",
							Description: "the id of target deck",
							Required:    true,
						},
					},
				},
				{
					Name:        "update",
					Description: "update exist deck",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "deck-id",
							Description: "the id of target deck",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "deck-name",
							Description: "name for the new deck",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "deck-description",
							Description: "description for the new deck",
							Required:    false,
						},
					},
				},
				{
					Name:        "delete",
					Description: "delete exist deck",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "deck-id",
							Description: "the id of target deck",
							Required:    true,
						},
					},
				},
			},
		},
	}
	slashCommandHandlerMap = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		deckCommandName: handleDeckCommand,
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
		authToken: quizdeck.Config.GetDiscordAuthToken(),
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

func createDeckEmbed(deck quizdeck.Deck, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       deck.Name,
		Description: deck.Description,
		Type:        discordgo.EmbedTypeRich,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID",
				Value:  strconv.FormatUint(uint64(deck.ID), 10),
				Inline: true,
			},
		},
	}
}

func (g *discordGateway) createSlashCommand() {

	for _, cmd := range commands {
		_, err := g.session.ApplicationCommandCreate(quizdeck.Config.GetDiscordAppID(), "", cmd)
		if err != nil {
			quizdeck.Logger.Fatal(
				"discord slash command create failed",
				zap.String("name", cmd.Name),
				zap.String("err", err.Error()),
			)
		}
	}
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	quizdeck.Logger.Info("recieve " + i.Data.Name)
	if handler, ok := slashCommandHandlerMap[i.Data.Name]; ok {
		handler(s, i)
	} else {
		quizdeck.Logger.Warn(
			"undefined slash commands",
			zap.String("command", i.Data.Name),
		)
	}
}

func handleDeckCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {

	content, embeds := "", []*discordgo.MessageEmbed{}
	switch i.Data.Options[0].Name {
	case "create":
		_content, _embeds := handleDeckCreateCommand(s, i)
		content, embeds = _content, append(embeds, _embeds...)
	case "list":
		embeds = append(embeds, handleDeckListCommand(s, i)...)
	case "show":
		embeds = append(embeds, handleDeckShowCommand(s, i)...)
	case "update":
		_content, _embeds := handleDeckUpdateCommand(s, i)
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

func handleDeckCreateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (string, []*discordgo.MessageEmbed) {
	name := i.Data.Options[0].Options[0].StringValue()
	desc := ""
	if len(i.Data.Options[0].Options) > 1 {
		desc = i.Data.Options[0].Options[1].StringValue()
	}
	userID := getDiscordUserID(i)
	cst, _ := quizdeck.DBConn.NewDeck(name, desc, userID, "")
	content := fmt.Sprintf("cassete %s created", name)
	embeds := []*discordgo.MessageEmbed{
		createDeckEmbed(cst, deckEmbedColorPublic),
	}
	return content, embeds
}

func handleDeckListCommand(s *discordgo.Session, i *discordgo.InteractionCreate) []*discordgo.MessageEmbed {

	userID := getDiscordUserID(i)
	embeds := []*discordgo.MessageEmbed{}
	csts, _ := quizdeck.DBConn.GetDecksByDiscordOwner(userID)
	for _, cst := range csts {
		embeds = append(embeds, createDeckEmbed(cst, deckEmbedColorPublic))
	}
	return embeds
}

func handleDeckShowCommand(s *discordgo.Session, i *discordgo.InteractionCreate) []*discordgo.MessageEmbed {

	id := i.Data.Options[0].Options[0].UintValue()
	cst, _ := quizdeck.DBConn.GetDeck(uint(id))
	embeds := []*discordgo.MessageEmbed{
		createDeckEmbed(cst, deckEmbedColorPublic),
	}
	return embeds
}

func handleDeckUpdateCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (string, []*discordgo.MessageEmbed) {

	id := i.Data.Options[0].Options[0].UintValue()
	mutations := map[string]interface{}{}
	if len(i.Data.Options[0].Options) > 1 {
		mutations["Name"] = i.Data.Options[0].Options[1].StringValue()
	}
	if len(i.Data.Options[0].Options) > 2 {
		mutations["Description"] = i.Data.Options[0].Options[2].StringValue()
	}
	cst, _ := quizdeck.DBConn.UpdateDeck(uint(id), mutations)
	content := fmt.Sprintf("deck %s updated", cst.Name)
	embeds := []*discordgo.MessageEmbed{
		createDeckEmbed(cst, deckEmbedColorPublic),
	}
	return content, embeds
}

func handleDeckDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) string {

	id := i.Data.Options[0].Options[0].UintValue()
	cst, _ := quizdeck.DBConn.GetDeck(uint(id))
	quizdeck.DBConn.DeleteDeck(uint(id))
	return fmt.Sprintf("deck %s deleted", cst.Name)
}

func (g *discordGateway) Start() {
	discordSession, _ := discordgo.New("Bot " + g.authToken)
	g.session = discordSession
	g.session.AddHandler(handleSlashCommand)
	g.session.Identify.Intents = discordgo.IntentsGuildMessages
	g.createSlashCommand()

	err := g.session.Open()
	if err != nil {
		quizdeck.Logger.Warn("error opening connection, " + err.Error())
		return
	}

	quizdeck.Logger.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	g.session.Close()
}
