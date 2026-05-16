package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nefu-dev/wx-note/internal/middleware"
	"github.com/nefu-dev/wx-note/internal/models"
	"golang.org/x/crypto/bcrypt"
)

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
	// 配置更新后清除微信客户端缓存，下次请求用新凭据重建
	h.wechatClients.Delete(uid)
	c.JSON(http.StatusOK, gin.H{"message": "配置已保存"})
}
