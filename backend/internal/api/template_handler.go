package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/models"
)

// ─── 模板 ─────────────────────────────────────────────────────

func (h *Handler) listTemplates(c *gin.Context) {
	uid := getUserID(c)
	category := c.Query("category")
	items, err := h.templateRepo.List(uid, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) createTemplate(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		Name      string `json:"name"`
		Category  string `json:"category"`
		Content   string `json:"content"`
		CoverURL  string `json:"cover_url"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	id, err := h.templateRepo.Create(&models.Template{
		UserID:    uid,
		Name:      req.Name,
		Category:  req.Category,
		Content:   req.Content,
		CoverURL:  req.CoverURL,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	t, _ := h.templateRepo.GetByID(id, uid)
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) getTemplate(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	t, err := h.templateRepo.GetByID(id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) updateTemplate(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")

	existing, err := h.templateRepo.GetByID(id, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		Category  string `json:"category"`
		Content   string `json:"content"`
		CoverURL  string `json:"cover_url"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	existing.Content = req.Content
	existing.CoverURL = req.CoverURL
	existing.SortOrder = req.SortOrder

	if err := h.templateRepo.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (h *Handler) deleteTemplate(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	if err := h.templateRepo.Delete(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func (h *Handler) getTemplateCategories(c *gin.Context) {
	uid := getUserID(c)
	categories, err := h.templateRepo.GetCategories(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *Handler) reorderTemplates(c *gin.Context) {
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

	if err := h.templateRepo.BatchUpdateSortOrder(items, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "排序失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "排序已保存"})
}
