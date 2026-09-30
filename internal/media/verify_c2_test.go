package media_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
)

func loadCriteria(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "acceptance", "media-package.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func runCandidate(t *testing.T, candidate []byte) kernel.RunResult {
	t.Helper()
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "c2", Text: "adversarial"},
		CriteriaJSON: loadCriteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: func(kernel.Objective, string) (kernel.Candidate, error) {
			return kernel.Candidate{Bytes: candidate, Producer: media.ProducerID}, nil
		},
		Transform:  media.Transform,
		Verify:     media.Verify,
		VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func validPackage() map[string]any {
	return map[string]any{
		"schema":    "tamvori.media.production_package.v1",
		"title":     "Valid Control",
		"audience":  "operators",
		"message":   "governed acceptance",
		"scenes":    []map[string]string{{"id": "1", "outline": "open"}},
		"narration": "A real narration line.",
		"assets":    []map[string]string{{"name": "a-roll", "hash": ""}},
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestC2ValidControlAccepted(t *testing.T) {
	res := runCandidate(t, mustJSON(t, validPackage()))
	if res.Status != "accepted" {
		t.Fatalf("status=%s reason=%s", res.Status, res.RejectReason)
	}
}

func TestC2EmptyTitleRejected(t *testing.T) {
	p := validPackage()
	p["title"] = ""
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("empty title accepted")
	}
}

func TestC2WhitespaceTitleRejected(t *testing.T) {
	p := validPackage()
	p["title"] = "   "
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("whitespace title accepted")
	}
}

func TestC2EmptyAssetNameRejected(t *testing.T) {
	p := validPackage()
	p["assets"] = []map[string]string{{"name": "", "hash": ""}}
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("empty asset name accepted")
	}
	if !strings.Contains(res.RejectReason, "empty asset name") && res.Status != "rejected" {
		t.Fatalf("unexpected reason %q status %s", res.RejectReason, res.Status)
	}
}

func TestC2WhitespaceAssetNameRejected(t *testing.T) {
	p := validPackage()
	p["assets"] = []map[string]string{{"name": "  ", "hash": ""}}
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("whitespace asset name accepted")
	}
}

func TestC2EmptySceneOutlineRejected(t *testing.T) {
	p := validPackage()
	p["scenes"] = []map[string]string{{"id": "1", "outline": ""}}
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("empty scene outline accepted")
	}
}

func TestC2WhitespaceSceneOutlineRejected(t *testing.T) {
	p := validPackage()
	p["scenes"] = []map[string]string{{"id": "1", "outline": "   "}}
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("whitespace scene outline accepted")
	}
}

func TestC2EmptySceneIDRejected(t *testing.T) {
	p := validPackage()
	p["scenes"] = []map[string]string{{"id": "", "outline": "open"}}
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("empty scene id accepted")
	}
}

func TestC2EmptyNarrationRejected(t *testing.T) {
	p := validPackage()
	p["narration"] = "   "
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("whitespace narration accepted")
	}
}

func TestC2PartialObjectRejected(t *testing.T) {
	res := runCandidate(t, []byte(`{"schema":"tamvori.media.production_package.v1","title":"Only Title"}`))
	if res.Status == "accepted" {
		t.Fatal("partial object accepted")
	}
}

func TestC2ModelSelfClaimIgnored(t *testing.T) {
	p := validPackage()
	p["title"] = "broken"
	p["valid"] = true
	p["passed"] = true
	p["approved"] = true
	res := runCandidate(t, mustJSON(t, p))
	if res.Status == "accepted" {
		t.Fatal("self-claim must not accept")
	}
}

func TestC2ExtraUnknownFieldsStillValid(t *testing.T) {
	p := validPackage()
	p["extra_noise"] = "ignored"
	res := runCandidate(t, mustJSON(t, p))
	if res.Status != "accepted" {
		t.Fatalf("extra fields should not block valid package: %s", res.RejectReason)
	}
}

func TestC2MalformedJSONRejected(t *testing.T) {
	res := runCandidate(t, []byte(`{not-json`))
	if res.Status == "accepted" {
		t.Fatal("malformed JSON accepted")
	}
}

func TestC2EmptyAssetNameFalseAcceptanceReproducedAgainstOldCriteria(t *testing.T) {
	// EXP-001C-era criteria lacked require_nonempty_asset_names.
	oldCriteria := []byte(`{
	  "schema": "tamvori.media.production_package.v1",
	  "required_fields": ["title", "audience", "message", "scenes", "narration", "assets"],
	  "require_sorted_scenes": true,
	  "require_sorted_assets": true,
	  "require_asset_hashes": true,
	  "min_scenes": 1,
	  "forbidden_titles": ["broken"]
	}`)
	candidate := mustJSON(t, map[string]any{
		"schema":    "tamvori.media.production_package.v1",
		"title":     "Looks Fine",
		"audience":  "ops",
		"message":   "msg",
		"scenes":    []map[string]string{{"id": "1", "outline": "open"}},
		"narration": "line",
		"assets":    []map[string]string{{"name": "", "hash": ""}},
	})
	out, err := media.Transform(candidate)
	if err != nil {
		t.Fatal(err)
	}
	old, err := media.Verify(out, oldCriteria, kernel.HashBytes(oldCriteria))
	if err != nil {
		t.Fatal(err)
	}
	if !old.Pass {
		t.Fatal("expected old criteria to falsely accept empty asset name")
	}
	newV, err := media.Verify(out, loadCriteria(t), kernel.HashBytes(loadCriteria(t)))
	if err != nil {
		t.Fatal(err)
	}
	if newV.Pass {
		t.Fatal("hardened criteria must reject empty asset name")
	}
}
