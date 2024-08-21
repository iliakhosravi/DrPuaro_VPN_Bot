package models

import (
	"net/url"

	"gorm.io/gorm"
)

type GuideType string

const (
	FWD_MSG_GUIDE GuideType = "forward_message"
	LINK_GUIDE    GuideType = "link"
)

type Guide struct {
	BaseModel
	Title    string    `json:"title"`
	Type     GuideType `json:"type"`
	FwdMsgID int       `json:"forward_message_id"`
	Link     string    `json:"link"`
}

func (guide *Guide) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Guide{})
}

func UrlValidator(link string) bool {
	parsedURL, err := url.ParseRequestURI(link)
	if err != nil {
		return false
	}

	// Check if the URL has a valid scheme (http or https)
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// Check if the URL has a host
	if parsedURL.Host == "" {
		return false
	}

	return true
}
