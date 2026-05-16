package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─── 标签 ─────────────────────────────────────────────────────

func (h *Handler) listTags(c *gin.Context) {
	uid := getUserID(c)
	tags, err := h.tagRepo.List(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *Handler) createTag(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：标签名称不能为空"})
		return
	}

	tag, err := h.tagRepo.Create(uid, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建标签失败"})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *Handler) deleteTag(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	if err := h.tagRepo.Delete(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ─── 素材 ─────────────────────────────────────────────────────

func (h *Handler) listMedia(c *gin.Context) {
	uid := getUserID(c)
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	items, total, err := h.mediaRepo.List(uid, page, pageSize)
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

func (h *Handler) deleteMedia(c *gin.Context) {
	uid := getUserID(c)
	id := getIntParam(c, "id")
	if err := h.mediaRepo.Delete(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
