package media_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
	"github.com/smcdaniel54/Tamvori/internal/openai"
)

func criteria(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "acceptance", "media-package.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func openaiEnvelope(content string) []byte {
	b, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"content": content}},
		},
	})
	return b
}

func TestLiveValidProviderResponseAccepted(t *testing.T) {
	pkg := `{
	  "schema":"tamvori.media.production_package.v1",
	  "title":"Live Draft",
	  "audience":"builders",
	  "message":"governed generation",
	  "scenes":[{"id":"2","outline":"end"},{"id":"1","outline":"start"}],
	  "narration":"A short line.",
	  "assets":[{"name":"b","hash":""},{"name":"a","hash":""}]
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(openaiEnvelope(pkg))
	}))
	defer srv.Close()

	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "live-1", Text: "demo package"},
		CriteriaJSON: criteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose:      media.ProposeLive(media.LiveConfig{APIKey: "test", Model: "gpt-4o-mini", Context: ctx}),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "accepted" {
		t.Fatalf("status=%s reason=%s", res.Status, res.RejectReason)
	}
	var sawProvider bool
	for _, e := range res.Events {
		if e.Kind == "provider_response" {
			sawProvider = true
			if !strings.Contains(e.Note, "provider=openai") {
				t.Fatalf("note=%q", e.Note)
			}
			if e.Actor == media.VerifierID {
				t.Fatal("provider actor must not be verifier")
			}
		}
	}
	if !sawProvider {
		t.Fatal("missing provider_response event")
	}
}

func TestLiveMalformedHTTPGovernedReject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "live-fail", Text: "x"},
		CriteriaJSON: criteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose:      media.ProposeLive(media.LiveConfig{APIKey: "test", Context: ctx, Timeout: 5 * time.Second}),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "accepted" {
		t.Fatal("must not accept on HTTP failure")
	}
}

func TestLiveMalformedCandidateRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(openaiEnvelope(`{"title":"only"}`))
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "live-bad", Text: "x"},
		CriteriaJSON: criteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose:      media.ProposeLive(media.LiveConfig{APIKey: "test", Context: ctx}),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "rejected" {
		t.Fatalf("want rejected got %s", res.Status)
	}
}

func TestLiveModelSelfApprovalIgnored(t *testing.T) {
	// Model claims valid:true and uses forbidden title — must still fail verification.
	pkg := `{"schema":"tamvori.media.production_package.v1","title":"broken","audience":"a","message":"m","scenes":[{"id":"1","outline":"o"}],"narration":"n","assets":[{"name":"a","hash":""}],"valid":true,"passed":true}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(openaiEnvelope(pkg))
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "live-claim", Text: "x"},
		CriteriaJSON: criteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose:      media.ProposeLive(media.LiveConfig{APIKey: "test", Context: ctx}),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "accepted" {
		t.Fatal("self-approval must not accept")
	}
}

func TestFrozenResponseReplayDeterministic(t *testing.T) {
	assistant := []byte(`{
	  "schema":"tamvori.media.production_package.v1",
	  "title":"Frozen Live",
	  "audience":"ops",
	  "message":"replay",
	  "scenes":[{"id":"1","outline":"one"}],
	  "narration":"Line.",
	  "assets":[{"name":"clip","hash":""}]
	}`)
	raw := openaiEnvelope(string(assistant))
	a, err := kernel.Execute(kernel.RunRequest{
		WorkDir: t.TempDir(), Objective: kernel.Objective{ID: "f1", Text: "t"},
		CriteriaJSON: criteria(t), Bounds: kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: media.ProposeFrozen(assistant, "gpt-4o-mini", raw),
		Transform: media.Transform, Verify: media.Verify, VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := kernel.Execute(kernel.RunRequest{
		WorkDir: t.TempDir(), Objective: kernel.Objective{ID: "f1", Text: "t"},
		CriteriaJSON: criteria(t), Bounds: kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: media.ProposeFrozen(assistant, "gpt-4o-mini", raw),
		Transform: media.Transform, Verify: media.Verify, VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != "accepted" || b.Status != "accepted" {
		t.Fatalf("status a=%s b=%s", a.Status, b.Status)
	}
	if a.ArtifactHash != b.ArtifactHash {
		t.Fatalf("replay hash mismatch")
	}
}

func TestCapturedLiveEvidenceReplays(t *testing.T) {
	assistant, err := os.ReadFile(filepath.Join("..", "..", "experiments", "exp-001c-provider-slice", "evidence", "frozen_assistant.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "experiments", "exp-001c-provider-slice", "evidence", "frozen_raw_response.json"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := kernel.Execute(kernel.RunRequest{
		WorkDir: t.TempDir(), Objective: kernel.Objective{ID: "capture", Text: "t"},
		CriteriaJSON: criteria(t), Bounds: kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: media.ProposeFrozen(assistant, "gpt-4o-mini", raw),
		Transform: media.Transform, Verify: media.Verify, VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Historical EXP-001C live package used empty asset names; hardened criteria reject it.
	if a.Status == "accepted" {
		t.Fatal("frozen 001C live response must no longer be accepted under nonempty asset-name criteria")
	}
	if !strings.Contains(a.RejectReason, "empty asset name") {
		t.Fatalf("expected empty asset name rejection, got %q", a.RejectReason)
	}
	b, err := kernel.Execute(kernel.RunRequest{
		WorkDir: t.TempDir(), Objective: kernel.Objective{ID: "capture", Text: "t"},
		CriteriaJSON: criteria(t), Bounds: kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: media.ProposeFrozen(assistant, "gpt-4o-mini", raw),
		Transform: media.Transform, Verify: media.Verify, VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != b.Status || a.RejectReason != b.RejectReason {
		t.Fatal("captured live evidence rejection must replay deterministically")
	}
}

func TestLiveTimeoutGovernedReject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write(openaiEnvelope(`{}`))
	}))
	defer srv.Close()
	ctx := openai.WithEndpoint(context.Background(), srv.URL)
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "live-to", Text: "x"},
		CriteriaJSON: criteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose:      media.ProposeLive(media.LiveConfig{APIKey: "test", Context: ctx, Timeout: 80 * time.Millisecond}),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status == "accepted" {
		t.Fatal("timeout must not accept")
	}
	if !strings.Contains(res.RejectReason, "propose:") {
		t.Fatalf("reason=%s", res.RejectReason)
	}
}
