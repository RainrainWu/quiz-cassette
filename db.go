package quiz_cassette

import (
	"fmt"

	"github.com/pkg/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	Connect() error
	NewCassette(name, desc, discordID, telegramID string) (Cassette, error)
	UpdateCassette(id uint, mutations map[string]interface{}) (Cassette, error)
	GetCassette(id uint) (Cassette, error)
	GetCassettesByDiscordOwner(discordID string) ([]Cassette, error)
	DeleteCassette(id uint) error
}

type database struct {
	connection *gorm.DB
	dsn        string
}

type Cassette struct {
	*gorm.Model
	Name            string        `json:"name" gorm:"not null"`
	Description     string        `json:"description" gorm:"not null"`
	OwnerDiscordID  string        `json:"owner_discord_id"`
	OwnerTelegramID string        `json:"owner_telegram_id"`
	Quizzes         []Quiz        `json:"quizzes" gorm:"foreignKey:ID"`
	Tags            []CassetteTag `json:"tags" gorm:"many2many:cassette_tag"`
}

type CassetteTag struct {
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
		&Cassette{},
		&CassetteTag{},
		&Quiz{},
	)
	return err
}

func (d *database) NewCassette(name, desc, discordID, telegramID string) (Cassette, error) {

	if discordID == "" && telegramID == "" {
		return Cassette{}, errors.New("no user id specified.")
	}
	record := Cassette{
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

func (d *database) UpdateCassette(id uint, mutations map[string]interface{}) (Cassette, error) {

	cassette, err := d.GetCassette(id)
	if err != nil {
		return cassette, err
	}
	d.connection.Model(&cassette).Updates(mutations)
	return cassette, nil
}

func (d *database) GetCassette(id uint) (Cassette, error) {

	var cassette Cassette
	if err := d.connection.First(&cassette, "ID = ?", id).Error; err != nil {
		return cassette, errors.Wrap(err, fmt.Sprintf("cassette with id %d not found", id))
	}
	return cassette, nil
}

func (d *database) GetCassettesByDiscordOwner(discordID string) ([]Cassette, error) {

	var cassettes []Cassette
	if err := d.connection.Where("owner_discord_id = ?", discordID).Find(&cassettes).Error; err != nil {
		return cassettes, errors.Wrap(err, fmt.Sprintf("no cassettes found with dicord user %s", discordID))
	}
	return cassettes, nil
}

func (d *database) DeleteCassette(id uint) error {

	cassette, err := d.GetCassette(id)
	if err != nil {
		return err
	}
	d.connection.Delete(&cassette, 1)
	return nil
}
