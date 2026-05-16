package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type MediaRepo struct {
	db *sqlx.DB
}

func NewMediaRepo(db *sqlx.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

// List 返回用户素材分页列表
func (r *MediaRepo) List(userID int64, page, pageSize int) ([]models.Media, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int
	err := r.db.Get(&total, "SELECT COUNT(*) FROM media WHERE user_id = ?", userID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var items []models.Media
	err = r.db.Select(&items,
		"SELECT * FROM media WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetByID 获取单个素材记录
func (r *MediaRepo) GetByID(id, userID int64) (*models.Media, error) {
	var m models.Media
	err := r.db.Get(&m, "SELECT * FROM media WHERE id = ? AND user_id = ?", id, userID)
	return &m, err
}

// Create 创建素材记录
func (r *MediaRepo) Create(media *models.Media) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO media (user_id, url, filename, size)
		VALUES (?, ?, ?, ?)`,
		media.UserID, media.URL, media.Filename, media.Size)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Delete 删除素材记录
func (r *MediaRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec("DELETE FROM media WHERE id = ? AND user_id = ?", id, userID)
	return err
}
