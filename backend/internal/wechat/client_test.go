package wechat

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test-app-id", "test-secret")
	if c.appID != "test-app-id" {
		t.Errorf("expected appID 'test-app-id', got %q", c.appID)
	}
	if c.appSecret != "test-secret" {
		t.Errorf("expected appSecret, got %q", c.appSecret)
	}
	if c.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestAPIError(t *testing.T) {
	e := &APIError{ErrCode: 40001, ErrMsg: "invalid credential"}
	if e.Error() != "[40001] invalid credential" {
		t.Errorf("unexpected error string: %q", e.Error())
	}
}

func TestPublishResult(t *testing.T) {
	r := &PublishResult{
		DraftMediaID: "test-media-id",
		Status:       "draft",
	}
	if r.DraftMediaID != "test-media-id" {
		t.Errorf("expected draft_media_id, got %q", r.DraftMediaID)
	}
	if r.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", r.Status)
	}
}
