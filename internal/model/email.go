package model

import (
	"github.com/jinzhu/gorm"
)

type Email struct {
	gorm.Model
	Remote    string
	IP        uint32
	From      string
	To        string
	Date      string
	EmailFrom string
	EmailTo   string
	Subject   string
	Raw       string
}
