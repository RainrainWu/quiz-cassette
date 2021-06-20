package quiz_cassette

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	Connect() error
}

type database struct {
	connection *gorm.DB
	dsn        string
}

type User struct {
	*gorm.Model
	DiscordUserID string
}

type Cassette struct {
	*gorm.Model
	Name        string        `json:"name" gorm:"not null"`
	Description string        `json:"description" gorm:"not null"`
	Quizzes     []Quiz        `json:"quizzes" gorm:"foreignKey:ID"`
	Tags        []CassetteTag `json:"tags" gorm:"many2many:cassette_tag;"`
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
		&User{},
		&Cassette{},
		&CassetteTag{},
		&Quiz{},
	)
	return err
}
