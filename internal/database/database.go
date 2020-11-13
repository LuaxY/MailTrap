package database

import (
	"log"

	"MailTrap/internal/config"
	"MailTrap/internal/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var cfg = config.Get()

var DB *gorm.DB

func init() {
	var err error

	DB, err = gorm.Open("sqlite3", cfg.Database)

	if err != nil {
		log.Fatal(err)
	}

	//defer DB.Close()

	DB.AutoMigrate(&model.Email{})
}
