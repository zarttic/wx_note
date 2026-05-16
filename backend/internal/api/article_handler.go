package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/models"
)

// ─── 文章 ─────────────────────────────────────────────────────

func (h *Handler) listArticles(c *gin.Context) {
	uid := getUserID(c)

	req := models.ArticleListRequest{
		Page:     getIntQuery(c, "page", 1),
		PageSize: getIntQuery(c, "page_size", 20),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
		TagID:    getInt64Query(c, "tag_id", 0),
	}

	result, err := h.articleRepo.List(req, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 批量填充标签
	if len(result.Items) > 0 {
		articleIDs := make([]int64, len(result.Items))
		for i, a := range result.Items {
			articleIDs[i] = a.ID
		}
		tagsMap, err := h.tagRepo.GetByArticleIDs(articleIDs)
		if err == nil {
			for i := range result.Items {
				result.Items[i].Tags = tagsMap[result.Items[i].ID]
				if result.Items[i].Tags == nil {
					result.Items[i].Tags = []models.Tag{}
				}
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) createArticle(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		Title    string  `json:"title"`
		Markdown string  `json:"markdown"`
		Summary  string  `json:"summary"`
		CoverURL string  `json:"cover_url"`
		Status   string  `json:"status"`
		TagIDs   []int64 `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}
	title := req.Title
	if title == "" {
		_, title = extractTitle(req.Markdown)
	}
	wordCount := len([]rune(req.Markdown))

	id, err := h.articleRepo.Create(&models.Article{
		UserID:    uid,
		Title:     title,
		Markdown:  req.Markdown,
		Summary:   req.Summary,
		CoverURL:  req.CoverURL,
		Status:    req.Status,
		WordCount: wordCount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 设置文章标签
	if len(req.TagIDs) > 0 {
		if err := h.tagRepo.SetArticleTags(id, req.TagIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置标签失败"})
			return
		}
	}

	article, _ := h.articleRepo.GetByID(id, uid)
	// 填充标签
	if article != nil {
		tags, err := h.tagRepo.GetByArticleID(id)
		if err == nil {
			article.Tags = tags
		}
	}
	c.JSON(http.StatusCreated, article)
}

func (h *Handler) getArticle(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	article, err := h.articleRepo.GetByID(id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}
	// 填充标签
	tags, err := h.tagRepo.GetByArticleID(id)
	if err == nil {
		article.Tags = tags
	}
	c.JSON(http.StatusOK, article)
}

func (h *Handler) updateArticle(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")

	existing, err := h.articleRepo.GetByID(id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	var req struct {
		Title    string  `json:"title"`
		Markdown string  `json:"markdown"`
		Summary  string  `json:"summary"`
		CoverURL string  `json:"cover_url"`
		Status   string  `json:"status"`
		TagIDs   []int64 `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Markdown != "" {
		existing.Markdown = req.Markdown
		existing.WordCount = len([]rune(req.Markdown))
	}
	if req.Summary != "" {
		existing.Summary = req.Summary
	}
	if req.CoverURL != "" {
		existing.CoverURL = req.CoverURL
	}
	if req.Status != "" {
		existing.Status = req.Status
	}

	if err := h.articleRepo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 设置文章标签（即使 tag_ids 为空也要清除已有标签）
	if err := h.tagRepo.SetArticleTags(id, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置标签失败"})
		return
	}

	// Create revision snapshot on update
	if _, err := h.revisionRepo.Create(&models.Revision{
		ArticleID: id,
		UserID:    uid,
		Title:     existing.Title,
		Markdown:  existing.Markdown,
		WordCount: existing.WordCount,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建版本记录失败"})
		return
	}
	// Keep only 30 most recent revisions per article
	_ = h.revisionRepo.DeleteOldRevisions(id, 30)

	// 填充标签
	tags, _ := h.tagRepo.GetByArticleID(id)
	existing.Tags = tags

	c.JSON(http.StatusOK, existing)
}

func (h *Handler) deleteArticle(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	if err := h.articleRepo.Delete(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *Handler) reorderArticles(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		Items []struct {
			ID        int64 `json:"id"`
			SortOrder int   `json:"sort_order"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	items := make([]struct {
		ID        int64
		SortOrder int
	}, len(req.Items))
	for i, r := range req.Items {
		items[i].ID = r.ID
		items[i].SortOrder = r.SortOrder
	}

	if err := h.articleRepo.BatchUpdateSortOrder(items, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "排序失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "排序已保存"})
}
