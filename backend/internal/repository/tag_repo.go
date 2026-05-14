package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/models"
)

type TagRepo struct {
	db *sqlx.DB
}

func NewTagRepo(db *sqlx.DB) *TagRepo {
	return &TagRepo{db: db}
}

// List 返回用户所有标签及文章数
func (r *TagRepo) List(userID int64) ([]models.TagWithCount, error) {
	var tags []models.TagWithCount
	err := r.db.Select(&tags,
		`SELECT t.*, COUNT(at.article_id) AS article_count
		 FROM tags t
		 LEFT JOIN article_tags at ON t.id = at.tag_id
		 WHERE t.user_id = ?
		 GROUP BY t.id
		 ORDER BY t.name`, userID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// Create 创建标签，若已存在则返回已有标签
func (r *TagRepo) Create(userID int64, name string) (*models.Tag, error) {
	// 先检查是否已存在
	var existing models.Tag
	err := r.db.Get(&existing,
		"SELECT * FROM tags WHERE user_id = ? AND name = ?", userID, name)
	if err == nil {
		return &existing, nil
	}

	result, err := r.db.Exec("INSERT INTO tags (user_id, name) VALUES (?, ?)", userID, name)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Tag{ID: id, UserID: userID, Name: name}, nil
}

// Delete 删除标签（级联删除 article_tags）
func (r *TagRepo) Delete(id, userID int64) error {
	_, err := r.db.Exec("DELETE FROM tags WHERE id = ? AND user_id = ?", id, userID)
	return err
}

// GetByArticleID 获取文章的标签
func (r *TagRepo) GetByArticleID(articleID int64) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Select(&tags,
		`SELECT t.* FROM tags t
		 INNER JOIN article_tags at ON t.id = at.tag_id
		 WHERE at.article_id = ?
		 ORDER BY t.name`, articleID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// SetArticleTags 设置文章标签（先删后插）
func (r *TagRepo) SetArticleTags(articleID int64, tagIDs []int64) error {
	// 先删除文章的所有标签关联
	_, err := r.db.Exec("DELETE FROM article_tags WHERE article_id = ?", articleID)
	if err != nil {
		return err
	}

	// 批量插入新关联
	for _, tagID := range tagIDs {
		_, err := r.db.Exec("INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?)", articleID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

// EnsureTags 批量确保标签存在，返回 tag IDs
func (r *TagRepo) EnsureTags(userID int64, names []string) ([]int64, error) {
	var ids []int64
	for _, name := range names {
		tag, err := r.Create(userID, name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, tag.ID)
	}
	return ids, nil
}