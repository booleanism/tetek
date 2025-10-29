package model

import "time"

type Feed struct {
	Id         string     `json:"id"`
	Title      string     `json:"title"`
	Url        string     `json:"url"`
	Text       string     `json:"text"`
	By         string     `json:"by"`
	Type       string     `json:"type"`
	Points     int        `json:"points"`
	NCommnents int        `json:"n_comments"`
	Created_At *time.Time `json:"created_at"`
	Deleted_At *time.Time `json:"deleted_at"`
}
