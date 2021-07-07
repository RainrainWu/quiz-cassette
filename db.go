package quiz_cassette

import (
	"fmt"

	"github.com/pkg/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	Connect() error
	NewDeck(name, desc, discordID, telegramID string) (Deck, error)
	UpdateDeck(id uint, mutations map[string]interface{}) (Deck, error)
	GetDeck(id uint) (Deck, error)
	GetDecksByDiscordOwner(discordID string) ([]Deck, error)
	DeleteDeck(id uint) error
}

type database struct {
	connection *gorm.DB
	dsn        string
}

type Deck struct {
	*gorm.Model
	Name            string    `json:"name" gorm:"not null"`
	Description     string    `json:"description" gorm:"not null"`
	OwnerDiscordID  string    `json:"owner_discord_id"`
	OwnerTelegramID string    `json:"owner_telegram_id"`
	Quizzes         []Quiz    `json:"quizzes" gorm:"foreignKey:ID"`
	Tags            []DeckTag `json:"tags" gorm:"many2many:deck_tag"`
}

type DeckTag struct {
	*gorm.Model
	Name  string `json:"name" gorm:"not null"`
	Count int    `json:"count" gorm:"not null"`
}

type Quiz struct {
	*gorm.Model
	Subject  string `json:"name"`
	Previous uint   `json:"previous" gorm:"not null"`
}

var (
	DBConn Database = NewDatabase()
)

func NewDatabase() Database {
	instance := database{
		connection: nil,
		dsn: fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			Config.GetPostgresHost(),
			Config.GetPostgresUser(),
			Config.GetPostgresPassword(),
			Config.GetPostgresDBName(),
			Config.GetPostgresPort(),
		),
	}
	instance.Connect()
	Logger.Info("postgresql database is ready.")
	return &instance
}

func (d *database) Connect() error {

	var err error
	d.connection, err = gorm.Open(postgres.Open(d.dsn), &gorm.Config{})
	d.connection.AutoMigrate(
		&Deck{},
		&DeckTag{},
		&Quiz{},
	)
	return err
}

func (d *database) NewDeck(name, desc, discordID, telegramID string) (Deck, error) {

	if discordID == "" && telegramID == "" {
		return Deck{}, errors.New("no user id specified.")
	}
	record := Deck{
		Name:        name,
		Description: desc,
	}
	if discordID != "" {
		record.OwnerDiscordID = discordID
	} else {
		record.OwnerTelegramID = telegramID
	}
	d.connection.Create(&record)
	return record, nil
}

func (d *database) UpdateDeck(id uint, mutations map[string]interface{}) (Deck, error) {

	deck, err := d.GetDeck(id)
	if err != nil {
		return deck, err
	}
	d.connection.Model(&deck).Updates(mutations)
	return deck, nil
}

func (d *database) GetDeck(id uint) (Deck, error) {

	var deck Deck
	if err := d.connection.First(&deck, "ID = ?", id).Error; err != nil {
		return deck, errors.Wrap(err, fmt.Sprintf("cassette with id %d not found", id))
	}
	return deck, nil
}

func (d *database) GetDecksByDiscordOwner(discordID string) ([]Deck, error) {

	var decks []Deck
	if err := d.connection.Where("owner_discord_id = ?", discordID).Find(&decks).Error; err != nil {
		return decks, errors.Wrap(err, fmt.Sprintf("no decks found with discord user %s", discordID))
	}
	return decks, nil
}

func (d *database) DeleteDeck(id uint) error {

	deck, err := d.GetDeck(id)
	if err != nil {
		return err
	}
	d.connection.Delete(&deck, 1)
	return nil
}
