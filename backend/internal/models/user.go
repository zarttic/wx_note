package models

type User struct {
	ID        int64  `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"-" db:"password"`
	Nickname  string `json:"nickname" db:"nickname"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type UserConfig struct {
	UserID        int64  `json:"-" db:"user_id"`
	WechatAppID   string `json:"wechat_app_id" db:"wechat_app_id"`
	WechatSecret  string `json:"wechat_secret" db:"wechat_secret"`
	DefaultAuthor string `json:"default_author" db:"default_author"`
	UpdatedAt     string `json:"updated_at" db:"updated_at"`
}

type UserConfigSafe struct {
	WechatAppID   string `json:"wechat_app_id"`
	HasSecret     bool   `json:"has_secret"`
	DefaultAuthor string `json:"default_author"`
}
