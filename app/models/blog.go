package models

import (
	"time"

	"gorm.io/gorm"
)

type Author struct {
	gorm.Model
	ID      string
	Name    string
	NameURL string
}

type BlogPost struct {
	gorm.Model
	ID           string
	AuthorID     string
	Author       Author
	Title        string
	Body         string
	CreationDate time.Time
}
