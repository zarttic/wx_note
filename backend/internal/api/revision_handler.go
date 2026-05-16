package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/models"
)

// ─── 版本历史 ────────────────────────────────────────────────────

func (h *Handler) listRevisions(c *gin.Context) {
	uid := getUserID(c)
	articleID := getIntParam(c, "id")
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	if _, err := h.articleRepo.GetByID(articleID, uid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	items, total, err := h.revisionRepo.List(articleID, uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"items": items,
	})
}

func (h *Handler) getRevision(c *gin.Context) {
	uid := getUserID(c)
	rid := getIntParam(c, "id")

	rev, err := h.revisionRepo.GetByID(rid, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "版本不存在"})
		return
	}
	c.JSON(http.StatusOK, rev)
}

func (h *Handler) restoreRevision(c *gin.Context) {
	uid := getUserID(c)
	articleID := getIntParam(c, "id")
	rid := getIntParam(c, "rid")

	existing, err := h.articleRepo.GetByID(articleID, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	rev, err := h.revisionRepo.GetByID(rid, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "版本不存在"})
		return
	}

	// Snapshot current state before restoring
	if _, err := h.revisionRepo.Create(&models.Revision{
		ArticleID: articleID,
		UserID:    uid,
		Title:     existing.Title,
		Markdown:  existing.Markdown,
		WordCount: existing.WordCount,
	}); err != nil {
		log.Printf("failed to snapshot current revision before restore: %v", err)
	}

	existing.Title = rev.Title
	existing.Markdown = rev.Markdown
	existing.WordCount = rev.WordCount
	if err := h.articleRepo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复失败"})
		return
	}

	tags, _ := h.tagRepo.GetByArticleID(articleID)
	existing.Tags = tags

	c.JSON(http.StatusOK, existing)
}
