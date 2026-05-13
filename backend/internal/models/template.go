package models

type Template struct {
	ID        int64  `json:"id" db:"id"`
	UserID    int64  `json:"-" db:"user_id"`
	Name      string `json:"name" db:"name"`
	Category  string `json:"category" db:"category"`
	Content   string `json:"content" db:"content"`
	CoverURL  string `json:"cover_url" db:"cover_url"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}
