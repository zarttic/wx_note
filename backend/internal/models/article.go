package models

type Article struct {
	ID           int64  `json:"id" db:"id"`
	UserID       int64  `json:"-" db:"user_id"`
	Title        string `json:"title" db:"title"`
	Markdown     string `json:"markdown" db:"markdown"`
	Summary      string `json:"summary" db:"summary"`
	CoverURL     string `json:"cover_url" db:"cover_url"`
	Status       string `json:"status" db:"status"`
	DraftMediaID string `json:"draft_media_id" db:"draft_media_id"`
	PublishID    int64  `json:"publish_id" db:"publish_id"`
	WordCount    int    `json:"word_count" db:"word_count"`
	CreatedAt    string `json:"created_at" db:"created_at"`
	UpdatedAt    string `json:"updated_at" db:"updated_at"`
	Tags         []Tag  `json:"tags"` // 非数据库字段，由 handler 层填充
}

type ArticleListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Search   string `form:"search"`
	TagID    int64  `form:"tag_id"`
}

type ArticleListResponse struct {
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Items []Article `json:"items"`
}
