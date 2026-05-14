package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type RevisionRepo struct {
	db *sqlx.DB
}

func NewRevisionRepo(db *sqlx.DB) *RevisionRepo {
	return &RevisionRepo{db: db}
}

func (r *RevisionRepo) Create(rev *models.Revision) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO article_revisions
		(article_id, user_id, title, markdown, word_count)
		VALUES (?, ?, ?, ?, ?)`,
		rev.ArticleID, rev.UserID, rev.Title, rev.Markdown, rev.WordCount,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *RevisionRepo) List(articleID, userID int64, page, pageSize int) ([]models.Revision, int, error) {
	var total int
	err := r.db.Get(&total, "SELECT COUNT(*) FROM article_revisions WHERE article_id = ? AND user_id = ?", articleID, userID)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items := make([]models.Revision, 0)
	err = r.db.Select(&items,
		"SELECT id, article_id, title, word_count, created_at FROM article_revisions WHERE article_id = ? AND user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		articleID, userID, pageSize, offset,
	)
	return items, total, err
}

func (r *RevisionRepo) GetByID(id, userID int64) (*models.Revision, error) {
	var rev models.Revision
	err := r.db.Get(&rev, "SELECT * FROM article_revisions WHERE id = ? AND user_id = ?", id, userID)
	return &rev, err
}

// DeleteOldRevisions keeps only the N most recent revisions for an article
func (r *RevisionRepo) DeleteOldRevisions(articleID int64, keep int) error {
	_, err := r.db.Exec(`DELETE FROM article_revisions WHERE article_id = ? AND id NOT IN (
		SELECT id FROM article_revisions WHERE article_id = ? ORDER BY created_at DESC LIMIT ?
	)`, articleID, articleID, keep)
	return err
}