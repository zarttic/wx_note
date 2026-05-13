package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type ArticleRepo struct {
	db *sqlx.DB
}

func NewArticleRepo(db *sqlx.DB) *ArticleRepo {
	return &ArticleRepo{db: db}
}

func (r *ArticleRepo) Create(a *models.Article) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO articles
		(user_id, title, markdown, summary, cover_url, status, word_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.UserID, a.Title, a.Markdown, a.Summary, a.CoverURL, a.Status, a.WordCount,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ArticleRepo) Update(a *models.Article) error {
	_, err := r.db.Exec(`UPDATE articles SET
		title = ?, markdown = ?, summary = ?, cover_url = ?, status = ?,
		draft_media_id = ?, publish_id = ?, word_count = ?, updated_at = datetime('now')
		WHERE id = ? AND user_id = ?`,
		a.Title, a.Markdown, a.Summary, a.CoverURL, a.Status,
		a.DraftMediaID, a.PublishID, a.WordCount, a.ID, a.UserID,
	)
	return err
}

func (r *ArticleRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec("DELETE FROM articles WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (r *ArticleRepo) GetByID(id, userID int64) (*models.Article, error) {
	var a models.Article
	err := r.db.Get(&a, "SELECT * FROM articles WHERE id = ? AND user_id = ?", id, userID)
	return &a, err
}

func (r *ArticleRepo) List(req models.ArticleListRequest, userID int64) (*models.ArticleListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	var conditions []string
	var args []interface{}
	conditions = append(conditions, "user_id = ?")
	args = append(args, userID)

	if req.Status != "" && req.Status != "all" {
		conditions = append(conditions, "status = ?")
		args = append(args, req.Status)
	}

	if req.Search != "" {
		conditions = append(conditions, "(title LIKE ? OR markdown LIKE ?)")
		args = append(args, "%"+req.Search+"%", "%"+req.Search+"%")
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := r.db.Get(&total, fmt.Sprintf("SELECT COUNT(*) FROM articles WHERE %s", where), countArgs...)
	if err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf("SELECT * FROM articles WHERE %s ORDER BY updated_at DESC LIMIT ? OFFSET ?", where)
	args = append(args, req.PageSize, offset)

	var items []models.Article
	if err := r.db.Select(&items, query, args...); err != nil {
		return nil, err
	}

	return &models.ArticleListResponse{
		Total: total,
		Page:  req.Page,
		Items: items,
	}, nil
}
