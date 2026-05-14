package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/middleware"
	"github.com/nefu-dev/wx-note/internal/models"
	"github.com/nefu-dev/wx-note/internal/repository"
	"github.com/nefu-dev/wx-note/internal/renderer"
	"github.com/nefu-dev/wx-note/internal/wechat"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db            *sqlx.DB
	userRepo      *repository.UserRepo
	articleRepo   *repository.ArticleRepo
	templateRepo  *repository.TemplateRepo
	tagRepo       *repository.TagRepo
	mediaRepo     *repository.MediaRepo
	revisionRepo  *repository.RevisionRepo
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		db:           db,
		userRepo:     repository.NewUserRepo(db),
		articleRepo:  repository.NewArticleRepo(db),
		templateRepo: repository.NewTemplateRepo(db),
		tagRepo:      repository.NewTagRepo(db),
		mediaRepo:    repository.NewMediaRepo(db),
			revisionRepo:  repository.NewRevisionRepo(db),
	}
}

func (h *Handler) Setup() *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 公开路由
	r.POST("/api/auth/register", h.register)
	r.POST("/api/auth/login", h.login)
	r.GET("/api/health", h.healthCheck)

	// 需要认证
	auth := r.Group("/api")
	auth.Use(middleware.AuthRequired())
	{
		auth.POST("/auth/logout", h.logout)

		auth.GET("/user/profile", h.getProfile)
		auth.PUT("/user/profile", h.updateProfile)
		auth.GET("/user/config", h.getConfig)
		auth.PUT("/user/config", h.updateConfig)

		auth.GET("/articles", h.listArticles)
		auth.POST("/articles", h.createArticle)
		auth.GET("/articles/:id", h.getArticle)
		auth.PUT("/articles/:id", h.updateArticle)
		auth.DELETE("/articles/:id", h.deleteArticle)
		auth.PUT("/articles/reorder", h.reorderArticles)
		auth.PUT("/templates/reorder", h.reorderTemplates)
		auth.GET("/articles/:id/revisions", h.listRevisions)
		auth.GET("/revisions/:id", h.getRevision)
		auth.POST("/articles/:id/revisions/:rid/restore", h.restoreRevision)

		auth.POST("/editor/preview", h.preview)
		auth.POST("/editor/upload-image", h.uploadImage)
		auth.POST("/editor/verify", h.verifyWechat)
		auth.POST("/editor/publish", h.publish)
		auth.GET("/themes", h.listThemes)
		auth.GET("/templates", h.listTemplates)
		auth.POST("/templates", h.createTemplate)
		auth.GET("/templates/categories/all", h.getTemplateCategories)
		auth.GET("/templates/:id", h.getTemplate)
		auth.PUT("/templates/:id", h.updateTemplate)
		auth.DELETE("/templates/:id", h.deleteTemplate)

		auth.GET("/tags", h.listTags)
		auth.POST("/tags", h.createTag)
		auth.DELETE("/tags/:id", h.deleteTag)

		auth.GET("/media", h.listMedia)
		auth.DELETE("/media/:id", h.deleteMedia)
	}

	return r
}

// ─── 认证 ─────────────────────────────────────────────────────

type registerReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	user, err := h.userRepo.Create(req.Username, string(hash), req.Nickname)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	})
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ─── 用户 ─────────────────────────────────────────────────────

func (h *Handler) getProfile(c *gin.Context) {
	uid := getUserID(c)
	user, err := h.userRepo.GetByID(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) updateProfile(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.userRepo.UpdateNickname(uid, req.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func (h *Handler) getConfig(c *gin.Context) {
	uid := getUserID(c)
	cfg, err := h.userRepo.GetConfig(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}
	safe := models.UserConfigSafe{
		WechatAppID:   cfg.WechatAppID,
		HasSecret:     cfg.WechatSecret != "",
		DefaultAuthor: cfg.DefaultAuthor,
		LastAuthor:    cfg.LastAuthor,
	}
	c.JSON(http.StatusOK, safe)
}

func (h *Handler) updateConfig(c *gin.Context) {
	uid := getUserID(c)
	var req struct {
		WechatAppID   string `json:"wechat_app_id"`
		WechatSecret  string `json:"wechat_secret"`
		DefaultAuthor string `json:"default_author"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := h.userRepo.UpdateConfig(uid, req.WechatAppID, req.WechatSecret, req.DefaultAuthor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "配置已保存"})
}

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

	// 为每个文章填充标签
	for i := range result.Items {
		tags, err := h.tagRepo.GetByArticleID(result.Items[i].ID)
		if err == nil {
			result.Items[i].Tags = tags
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
	h.revisionRepo.Create(&models.Revision{
		ArticleID: id,
		UserID:    uid,
		Title:     existing.Title,
		Markdown:  existing.Markdown,
		WordCount: existing.WordCount,
	})
	// Keep only 30 most recent revisions per article
	h.revisionRepo.DeleteOldRevisions(id, 30)
	// 填充标签
	tags, err := h.tagRepo.GetByArticleID(id)
	if err == nil {
		existing.Tags = tags
	}

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
	cfg, err := h.userRepo.GetConfig(uid)
	if err != nil || cfg.WechatAppID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置微信公众号"})
		return
	}

	client := wechat.NewClient(cfg.WechatAppID, cfg.WechatSecret)
	url, err := client.UploadImage(tmpFile.Name())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("上传失败: %v", err)})
		return
	}

	// 上传成功后记录到素材库
	h.mediaRepo.Create(&models.Media{
		UserID:   uid,
		URL:      url,
		Filename: header.Filename,
		Size:     header.Size,
	})

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *Handler) verifyWechat(c *gin.Context) {
	uid := getUserID(c)
	cfg, err := h.userRepo.GetConfig(uid)
	if err != nil || cfg.WechatAppID == "" || cfg.WechatSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置微信公众号 AppID 和 AppSecret"})
		return
	}

	client := wechat.NewClient(cfg.WechatAppID, cfg.WechatSecret)
	if _, err := client.GetToken(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("验证失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

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
	tmpFile.ReadFrom(coverFile)
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
	h.userRepo.SaveLastAuthor(uid, author)

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
				h.articleRepo.Update(existing)

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
	h.articleRepo.Create(&models.Article{
		UserID:       uid,
		Title:        title,
		Markdown:     markdown,
		Summary:      digest,
		Status:       "published",
		DraftMediaID: result.DraftMediaID,
		WordCount:    wordCount,
	})

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"title":          title,
		"draft_media_id": result.DraftMediaID,
		"status":         "published",
	})
}

// ─── 主题 ───────────────────────────────────────────────────────

func (h *Handler) listThemes(c *gin.Context) {
	themes := renderer.AvailableThemes()
	c.JSON(http.StatusOK, gin.H{"themes": themes})
}

// ─── 工具函数 ─────────────────────────────────────────────────

func getUserID(c *gin.Context) int64 {
	s, _ := c.Get("user_id")
	id, _ := strconv.ParseInt(s.(string), 10, 64)
	return id
}

func getIntQuery(c *gin.Context, key string, defaultVal int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}

func getIntParam(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return v
}

func getInt64Query(c *gin.Context, key string, defaultVal int64) int64 {
	v, err := strconv.ParseInt(c.Query(key), 10, 64)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}

func extractTitle(md string) (title, body string) {
	lines := strings.Split(strings.TrimSpace(md), "\n")
	title = "无标题"
	start := 0
	for i, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "# ") {
			title = strings.TrimSpace(s[2:])
			start = i + 1
			break
		}
		if strings.HasPrefix(s, "## ") {
			title = strings.TrimSpace(s[3:])
			start = i + 1
			break
		}
		title = s
		start = i + 1
		break
	}
	body = strings.TrimSpace(strings.Join(lines[start:], "\n"))
	return
}

func extractSummary(md string) string {
	for _, line := range strings.Split(md, "\n") {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "![") {
			continue
		}
		// Strip basic markdown syntax for a plain-text summary
		s = strings.ReplaceAll(s, "**", "")
		s = strings.ReplaceAll(s, "*", "")
		s = strings.ReplaceAll(s, "`", "")
		runes := []rune(s)
		if len(runes) > 120 {
			return string(runes[:120])
		}
		return s
	}
	return ""
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

// ─── 排序 ─────────────────────────────────────────────────────

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
	for _, item := range req.Items {
		if err := h.templateRepo.UpdateSortOrder(item.ID, uid, item.SortOrder); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "排序失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "排序已保存"})
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
	for _, item := range req.Items {
		if err := h.articleRepo.UpdateSortOrder(item.ID, uid, item.SortOrder); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "排序失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "排序已保存"})
}

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
	h.revisionRepo.Create(&models.Revision{
		ArticleID: articleID,
		UserID:    uid,
		Title:     existing.Title,
		Markdown:  existing.Markdown,
		WordCount: existing.WordCount,
	})

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
