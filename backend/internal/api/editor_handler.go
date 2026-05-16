package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/renderer"
)

// ─── 编辑器 ───────────────────────────────────────────────────

type previewReq struct {
	Markdown string `json:"markdown"`
	Theme    string `json:"theme"`
}

type previewResp struct {
	HTML    string `json:"html"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func (h *Handler) preview(c *gin.Context) {
	var req previewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	theme := req.Theme
	if theme == "" {
		theme = "default"
	}

	title, body := extractTitle(req.Markdown)
	html := renderer.RenderWithTheme(body, theme)
	summary := extractSummary(body)

	c.JSON(http.StatusOK, previewResp{
		HTML:    html,
		Title:   title,
		Summary: summary,
	})
}

// ─── 主题 ───────────────────────────────────────────────────────

func (h *Handler) listThemes(c *gin.Context) {
	themes := renderer.AvailableThemes()
	c.JSON(http.StatusOK, gin.H{"themes": themes})
}
