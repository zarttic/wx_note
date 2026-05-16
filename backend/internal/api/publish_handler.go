package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/models"
	"github.com/nefu-dev/wx-note/internal/renderer"
	"github.com/nefu-dev/wx-note/internal/wechat"
)

// ─── 上传图片 ──────────────────────────────────────────────────

func (h *Handler) uploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择图片"})
		return
	}
	defer file.Close()

	suffix := filepath.Ext(header.Filename)
	if suffix == "" {
		suffix = ".jpg"
	}

	tmpFile, err := os.CreateTemp("", "upload-*"+suffix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.ReadFrom(file); err != nil {
		tmpFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	tmpFile.Close()

	uid := getUserID(c)
	client, err := h.getWechatClient(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := client.UploadImage(tmpFile.Name())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("上传失败: %v", err)})
		return
	}

	// 上传成功后记录到素材库
	if _, err := h.mediaRepo.Create(&models.Media{
		UserID:   uid,
		URL:      url,
		Filename: header.Filename,
		Size:     header.Size,
	}); err != nil {
		log.Printf("failed to record media in library: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// ─── 验证微信连接 ──────────────────────────────────────────────

func (h *Handler) verifyWechat(c *gin.Context) {
	uid := getUserID(c)
	client, err := h.getWechatClient(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := client.GetToken(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("验证失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

// ─── 发布 ─────────────────────────────────────────────────────

func (h *Handler) publish(c *gin.Context) {
	uid := getUserID(c)
	cfg, err := h.userRepo.GetConfig(uid)
	if err != nil || cfg.WechatAppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置微信公众号"})
		return
	}

	markdown := c.PostForm("markdown")
	if strings.TrimSpace(markdown) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章内容不能为空"})
		return
	}

	coverFile, header, err := c.Request.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择封面图片"})
		return
	}
	defer coverFile.Close()

	suffix := filepath.Ext(header.Filename)
	if suffix == "" {
		suffix = ".jpg"
	}

	tmpFile, err := os.CreateTemp("", "cover-*"+suffix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	coverPath := tmpFile.Name()
	if _, err := tmpFile.ReadFrom(coverFile); err != nil {
		tmpFile.Close()
		os.Remove(coverPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存封面失败"})
		return
	}
	tmpFile.Close()
	defer os.Remove(coverPath)

	author := c.PostForm("author")
	if author == "" {
		author = cfg.DefaultAuthor
	}

	theme := c.PostForm("theme")
	if theme == "" {
		theme = "default"
	}

	title, body := extractTitle(markdown)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文章标题不能为空"})
		return
	}

	digest := extractSummary(body)
	htmlContent := renderer.RenderWithTheme(body, theme)

	client := wechat.NewClient(cfg.WechatAppID, cfg.WechatSecret)
	result, err := client.PublishArticle(title, htmlContent, coverPath, author, digest)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("发布失败: %v", err)})
		return
	}

	// 保存本次发布使用的作者名，下次发布时自动回填
	_ = h.userRepo.SaveLastAuthor(uid, author)

	// 如果传了 article_id，更新已有文章；否则创建新文章
	articleIDStr := c.PostForm("article_id")
	if articleIDStr != "" {
		articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
		if err == nil {
			existing, err := h.articleRepo.GetByID(articleID, uid)
			if err == nil {
				existing.Status = "published"
				existing.DraftMediaID = result.DraftMediaID
				existing.Title = title
				existing.Markdown = markdown
				existing.Summary = digest
				existing.WordCount = len([]rune(markdown))
				_ = h.articleRepo.Update(existing)

				// Fill tags for response
				tags, _ := h.tagRepo.GetByArticleID(articleID)
				existing.Tags = tags

				c.JSON(http.StatusOK, gin.H{
					"ok":             true,
					"title":          title,
					"draft_media_id": result.DraftMediaID,
					"status":         "published",
					"article_id":     articleID,
				})
				return
			}
		}
	}

	// No article_id or article not found — create new
	wordCount := len([]rune(markdown))
	if _, err := h.articleRepo.Create(&models.Article{
		UserID:       uid,
		Title:        title,
		Markdown:     markdown,
		Summary:      digest,
		Status:       "published",
		DraftMediaID: result.DraftMediaID,
		WordCount:    wordCount,
	}); err != nil {
		log.Printf("failed to create article record after publish: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"title":          title,
		"draft_media_id": result.DraftMediaID,
		"status":         "published",
	})
}
