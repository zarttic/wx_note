package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const apiBase = "https://api.weixin.qq.com/cgi-bin"

type APIError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.ErrCode, e.ErrMsg)
}

type Client struct {
	appID      string
	appSecret  string
	token      string
	tokenExp   time.Time
	mu         sync.Mutex
	httpClient *http.Client
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) GetToken() (string, error) {
	return c.getToken(false)
}

func (c *Client) UploadImage(imagePath string) (string, error) {
	token, err := c.getToken(false)
	if err != nil {
		return "", err
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("media", filepath.Base(imagePath))
	io.Copy(part, file)
	writer.Close()

	url := fmt.Sprintf("%s/media/uploadimg?access_token=%s", apiBase, token)
	resp, err := c.httpClient.Post(url, writer.FormDataContentType(), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
		*APIError
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.URL == "" {
		return "", &APIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	return result.URL, nil
}

func (c *Client) getToken(forceRefresh bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !forceRefresh && c.token != "" && time.Now().Before(c.tokenExp.Add(-5*time.Minute)) {
		return c.token, nil
	}

	url := fmt.Sprintf("%s/token?grant_type=client_credential&appid=%s&secret=%s",
		apiBase, c.appID, c.appSecret)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		*APIError
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	if data.AccessToken == "" {
		return "", &APIError{ErrCode: data.ErrCode, ErrMsg: data.ErrMsg}
	}

	c.token = data.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(data.ExpiresIn) * time.Second)
	return c.token, nil
}

func (c *Client) UploadThumb(imagePath string) (string, error) {
	token, err := c.getToken(false)
	if err != nil {
		return "", err
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("media", filepath.Base(imagePath))
	io.Copy(part, file)
	writer.WriteField("type", "thumb")
	writer.Close()

	url := fmt.Sprintf("%s/material/add_material?access_token=%s", apiBase, token)
	resp, err := c.httpClient.Post(url, writer.FormDataContentType(), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		MediaID string `json:"media_id"`
		*APIError
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.MediaID == "" {
		return "", &APIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	return result.MediaID, nil
}

func (c *Client) AddDraft(title, content, thumbMediaID, author, digest string) (string, error) {
	token, err := c.getToken(false)
	if err != nil {
		return "", err
	}

	article := map[string]any{
		"title":           title,
		"content":         content,
		"thumb_media_id":  thumbMediaID,
	}
	if author != "" {
		article["author"] = author
	}
	if digest != "" {
		article["digest"] = digest
	}

	body := map[string]any{"articles": []any{article}}
	data, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/draft/add?access_token=%s", apiBase, token)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		MediaID string `json:"media_id"`
		*APIError
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.MediaID == "" {
		return "", &APIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	return result.MediaID, nil
}

type PublishResult struct {
	DraftMediaID string `json:"draft_media_id"`
	Status       string `json:"status"`
}

func (c *Client) PublishArticle(title, htmlContent, coverPath, author, digest string) (*PublishResult, error) {
	thumbID, err := c.UploadThumb(coverPath)
	if err != nil {
		return nil, fmt.Errorf("upload cover: %w", err)
	}

	draftID, err := c.AddDraft(title, htmlContent, thumbID, author, digest)
	if err != nil {
		return nil, fmt.Errorf("add draft: %w", err)
	}

	return &PublishResult{
		DraftMediaID: draftID,
		Status:       "draft",
	}, nil
}
