package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type TemplateRepo struct {
	db *sqlx.DB
}

func NewTemplateRepo(db *sqlx.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

func (r *TemplateRepo) Create(t *models.Template) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO templates
		(user_id, name, category, content, cover_url, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.UserID, t.Name, t.Category, t.Content, t.CoverURL, t.SortOrder,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TemplateRepo) Update(t *models.Template) error {
	_, err := r.db.Exec(`UPDATE templates SET
		name = ?, category = ?, content = ?, cover_url = ?, sort_order = ?, updated_at = datetime('now')
		WHERE id = ? AND user_id = ?`,
		t.Name, t.Category, t.Content, t.CoverURL, t.SortOrder, t.ID, t.UserID,
	)
	return err
}

func (r *TemplateRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec("DELETE FROM templates WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (r *TemplateRepo) GetByID(id, userID int64) (*models.Template, error) {
	var t models.Template
	err := r.db.Get(&t, "SELECT * FROM templates WHERE id = ? AND user_id = ?", id, userID)
	return &t, err
}

func (r *TemplateRepo) List(userID int64, category string) ([]models.Template, error) {
	var items []models.Template
	var err error
	if category != "" {
		err = r.db.Select(&items, "SELECT * FROM templates WHERE user_id = ? AND category = ? ORDER BY sort_order ASC, id ASC", userID, category)
	} else {
		err = r.db.Select(&items, "SELECT * FROM templates WHERE user_id = ? ORDER BY sort_order ASC, id ASC", userID)
	}
	return items, err
}

func (r *TemplateRepo) GetCategories(userID int64) ([]string, error) {
	var categories []string
	err := r.db.Select(&categories, "SELECT DISTINCT category FROM templates WHERE user_id = ? ORDER BY category ASC", userID)
	return categories, err
}
