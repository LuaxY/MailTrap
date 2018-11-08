package model

import (
	"github.com/jinzhu/gorm"
)

type Email struct {
	gorm.Model
	Remote    string
	IP        int
	From      string
	To        string
	Date      string
	EmailFrom string
	EmailTo   string
	Subject   string
	Body      string
}
