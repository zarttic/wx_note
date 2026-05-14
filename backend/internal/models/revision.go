package models

type Revision struct {
	ID        int64  `json:"id" db:"id"`
	ArticleID int64  `json:"article_id" db:"article_id"`
	UserID    int64  `json:"-" db:"user_id"`
	Title     string `json:"title" db:"title"`
	Markdown  string `json:"markdown" db:"markdown"`
	WordCount int    `json:"word_count" db:"word_count"`
	CreatedAt string `json:"created_at" db:"created_at"`
}