package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/openai"
)

func TestChatJSONValid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": `{"title":"ok"}`}},
			},
		})
	}))
	defer srv.Close()

	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	content, raw, err := openai.ChatJSON(ctx, "test-key", "gpt-4o-mini", "sys", "user", 100)
	if err != nil {
		t.Fatal(err)
	}
	if content != `{"title":"ok"}` {
		t.Fatalf("content=%q", content)
	}
	if len(raw) == 0 {
		t.Fatal("expected raw body")
	}
}

func TestChatJSONHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	_, raw, err := openai.ChatJSON(ctx, "test-key", "gpt-4o-mini", "s", "u", 50)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "http 502") {
		t.Fatalf("err=%v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected raw error body")
	}
}

func TestChatJSONMalformedProviderJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	_, _, err := openai.ChatJSON(ctx, "k", "m", "s", "u", 50)
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("err=%v", err)
	}
}

func TestChatJSONTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": `{}`}}},
		})
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(openai.WithEndpoint(context.Background(), srv.URL), 50*time.Millisecond)
	defer cancel()
	_, _, err := openai.ChatJSON(ctx, "k", "m", "s", "u", 50)
	if err == nil {
		t.Fatal("expected timeout")
	}
}
