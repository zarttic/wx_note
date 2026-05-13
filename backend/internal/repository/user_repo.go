package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(username, passwordHash, nickname string) (*models.User, error) {
	result, err := r.db.Exec(
		"INSERT INTO users (username, password, nickname) VALUES (?, ?, ?)",
		username, passwordHash, nickname,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return r.GetByID(id)
}

func (r *UserRepo) GetByID(id int64) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	return &user, err
}

func (r *UserRepo) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE username = ?", username)
	return &user, err
}

func (r *UserRepo) UpdateNickname(id int64, nickname string) error {
	_, err := r.db.Exec(
		"UPDATE users SET nickname = ?, updated_at = datetime('now') WHERE id = ?",
		nickname, id,
	)
	return err
}

func (r *UserRepo) GetConfig(userID int64) (*models.UserConfig, error) {
	var cfg models.UserConfig
	err := r.db.Get(&cfg, "SELECT * FROM user_configs WHERE user_id = ?", userID)
	if err != nil {
		// 不存在则创建默认配置
		_, _ = r.db.Exec("INSERT OR IGNORE INTO user_configs (user_id) VALUES (?)", userID)
		err = r.db.Get(&cfg, "SELECT * FROM user_configs WHERE user_id = ?", userID)
	}
	return &cfg, err
}

func (r *UserRepo) UpdateConfig(userID int64, appID, secret, author string) error {
	_, err := r.db.Exec(`UPDATE user_configs SET
		wechat_app_id = ?, wechat_secret = ?, default_author = ?, updated_at = datetime('now')
		WHERE user_id = ?`,
		appID, secret, author, userID,
	)
	return err
}
