package models

type Tag struct {
	ID     int64  `json:"id" db:"id"`
	UserID int64  `json:"-" db:"user_id"`
	Name   string `json:"name" db:"name"`
}

type TagWithCount struct {
	Tag
	ArticleCount int `json:"article_count" db:"article_count"`
}
