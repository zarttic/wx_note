package api

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/nefu-dev/wx-note/internal/middleware"
	"github.com/nefu-dev/wx-note/internal/repository"
	"github.com/nefu-dev/wx-note/internal/wechat"
)

// Handler holds all the HTTP handler state and dependencies.
type Handler struct {
	db            *sqlx.DB
	userRepo      *repository.UserRepo
	articleRepo   *repository.ArticleRepo
	templateRepo  *repository.TemplateRepo
	tagRepo       *repository.TagRepo
	mediaRepo     *repository.MediaRepo
	revisionRepo  *repository.RevisionRepo
	wechatClients sync.Map // map[int64]*wechat.Client
}

// NewHandler creates a new Handler with the given database connection.
func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		db:           db,
		userRepo:     repository.NewUserRepo(db),
		articleRepo:  repository.NewArticleRepo(db),
		templateRepo: repository.NewTemplateRepo(db),
		tagRepo:      repository.NewTagRepo(db),
		mediaRepo:    repository.NewMediaRepo(db),
		revisionRepo: repository.NewRevisionRepo(db),
	}
}

// getWechatClient returns a cached wechat.Client for the given user,
// or creates one from the user's stored config. The cache entry is
// cleaned up after 30 minutes.
func (h *Handler) getWechatClient(uid int64) (*wechat.Client, error) {
	if v, ok := h.wechatClients.Load(uid); ok {
		return v.(*wechat.Client), nil
	}

	cfg, err := h.userRepo.GetConfig(uid)
	if err != nil || cfg.WechatAppID == "" || cfg.WechatSecret == "" {
		return nil, fmt.Errorf("请先配置微信公众号")
	}

	client := wechat.NewClient(cfg.WechatAppID, cfg.WechatSecret)
	h.wechatClients.Store(uid, client)

	go func() {
		time.Sleep(30 * time.Minute)
		h.wechatClients.Delete(uid)
	}()

	return client, nil
}

// allowedOrigin returns the CORS origin based on the CORS_ORIGIN env var.
// Defaults to "*" for development; set to a specific origin in production.
func allowedOrigin() string {
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		return "*"
	}
	return origin
}

// Setup configures all routes and returns a gin.Engine ready to serve.
func (h *Handler) Setup() *gin.Engine {
	r := gin.Default()

	orig := allowedOrigin()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", orig)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
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
