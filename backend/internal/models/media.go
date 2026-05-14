package models

type Media struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"-" db:"user_id"`
	URL       string `json:"url" db:"url"`
	Filename  string `json:"filename" db:"filename"`
	Size      int64  `json:"size" db:"size"`
	CreatedAt string `json:"created_at" db:"created_at"`
}