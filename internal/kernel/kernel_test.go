package kernel_test

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
	path := filepath.Join("..", "..", "experiments", "exp-001b-kernel-slice", "acceptance", "criteria.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read criteria: %v", err)
	}
	return b
}

func runWith(t *testing.T, mode string, dir string) kernel.RunResult {
	t.Helper()
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      dir,
		Objective:    kernel.Objective{ID: "obj-test", Text: "kernel slice test"},
		CriteriaJSON: loadCriteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 2, MaxBytes: 1 << 20},
		Propose:      media.Propose(mode),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return res
}

func underDir(dir, writePath string) bool {
	absDir, err1 := filepath.Abs(dir)
	absPath, err2 := filepath.Abs(writePath)
	if err1 != nil || err2 != nil {
		return false
	}
	sep := string(filepath.Separator)
	prefix := absDir
	if !strings.HasSuffix(prefix, sep) {
		prefix += sep
	}
	return absPath == absDir || strings.HasPrefix(absPath, prefix)
}

func TestValidRunAccepted(t *testing.T) {
	res := runWith(t, media.FixtureValid, t.TempDir())
	if res.Status != "accepted" {
		t.Fatalf("status=%s reason=%s", res.Status, res.RejectReason)
	}
	if res.ArtifactHash == "" {
		t.Fatal("missing artifact hash")
	}
	var sawVerdictPass bool
	for _, e := range res.Events {
		if e.Kind == "verdict" && e.Result == "pass" {
			if e.Actor == media.ProducerID {
				t.Fatal("pass verdict from producer")
			}
			if e.Actor != media.VerifierID {
				t.Fatalf("unexpected verifier %s", e.Actor)
			}
			sawVerdictPass = true
		}
	}
	if !sawVerdictPass {
		t.Fatal("no independent pass verdict")
	}
}

func TestInvalidRunRejected(t *testing.T) {
	res := runWith(t, media.FixtureInvalid, t.TempDir())
	if res.Status != "rejected" {
		t.Fatalf("want rejected, got %s", res.Status)
	}
}

func TestWorkerSelfApprovalIgnored(t *testing.T) {
	res := runWith(t, media.FixtureClaimPass, t.TempDir())
	if res.Status == "accepted" {
		t.Fatal("claim_pass must not be accepted")
	}
	var sawClaim bool
	for _, e := range res.Events {
		if e.Kind == "claim" {
			sawClaim = true
		}
	}
	if !sawClaim {
		t.Fatal("expected claim event to be recorded")
	}
	if res.Projection.Status == "accepted" {
		t.Fatal("projection accepted from claim")
	}
}

func TestPinnedCriteriaImmutable(t *testing.T) {
	criteria := loadCriteria(t)
	pinned := kernel.HashBytes(criteria)
	dir := t.TempDir()
	acceptance := filepath.Join(dir, "acceptance")
	if err := os.MkdirAll(acceptance, 0o755); err != nil {
		t.Fatal(err)
	}
	critPath := filepath.Join(acceptance, "criteria.json")
	if err := os.WriteFile(critPath, criteria, 0o644); err != nil {
		t.Fatal(err)
	}

	workerWrite := filepath.Join(acceptance, "hacked.json")
	if !underDir(acceptance, workerWrite) {
		t.Fatal("expected worker write path under acceptance dir")
	}
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      filepath.Join(dir, "run"),
		Objective:    kernel.Objective{ID: "obj-pin", Text: "pin test"},
		CriteriaJSON: criteria,
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: func(obj kernel.Objective, planHash string) (kernel.Candidate, error) {
			_ = os.WriteFile(critPath, []byte(`{"schema":"tampered","required_fields":[]}`), 0o644)
			return media.Propose(media.FixtureValid)(obj, planHash)
		},
		Transform:  media.Transform,
		Verify:     media.Verify,
		VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.CriteriaHash != pinned {
		t.Fatalf("criteria hash changed: got %s want %s", res.CriteriaHash, pinned)
	}
	disk, _ := os.ReadFile(critPath)
	if kernel.HashBytes(disk) == pinned {
		t.Fatal("expected on-disk criteria to be tampered for the test setup")
	}
	if res.Status != "accepted" {
		t.Fatalf("run should still pass under pinned criteria, got %s (%s)", res.Status, res.RejectReason)
	}
}

func TestReplaySameArtifactHash(t *testing.T) {
	a := runWith(t, media.FixtureValid, t.TempDir())
	b := runWith(t, media.FixtureValid, t.TempDir())
	if a.ArtifactHash != b.ArtifactHash {
		t.Fatalf("replay mismatch %s vs %s", a.ArtifactHash, b.ArtifactHash)
	}
	if a.PlanHash != b.PlanHash {
		t.Fatalf("plan hash mismatch")
	}
	if a.CriteriaHash != b.CriteriaHash {
		t.Fatalf("criteria hash mismatch")
	}
}

func TestContentAddressChangesWithBytes(t *testing.T) {
	h1 := kernel.HashBytes([]byte(`{"a":1}`))
	h2 := kernel.HashBytes([]byte(`{"a":2}`))
	if h1 == h2 {
		t.Fatal("different bytes must different hashes")
	}
	if h1 != kernel.HashBytes([]byte(`{"a":1}`)) {
		t.Fatal("same bytes must same hash")
	}
	dir := t.TempDir()
	res := runWith(t, media.FixtureValid, dir)
	stored := filepath.Join(dir, "artifacts", res.ArtifactHash)
	b, err := os.ReadFile(stored)
	if err != nil {
		t.Fatal(err)
	}
	if kernel.HashBytes(b) != res.ArtifactHash {
		t.Fatal("stored artifact hash mismatch")
	}
}

func TestDeterministicTransform(t *testing.T) {
	raw, err := media.Propose(media.FixtureValid)(kernel.Objective{ID: "x", Text: "y"}, "plan")
	if err != nil {
		t.Fatal(err)
	}
	o1, err := media.Transform(raw.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	o2, err := media.Transform(raw.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(o1) != string(o2) {
		t.Fatal("transform not deterministic")
	}
	if kernel.HashBytes(o1) != kernel.HashBytes(o2) {
		t.Fatal("hash mismatch")
	}
	var pkg struct {
		Scenes []struct {
			ID string `json:"id"`
		} `json:"scenes"`
	}
	if err := json.Unmarshal(o1, &pkg); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(pkg.Scenes); i++ {
		if pkg.Scenes[i-1].ID > pkg.Scenes[i].ID {
			t.Fatal("scenes not sorted")
		}
	}
}

func TestProvenanceCapturesEvidence(t *testing.T) {
	res := runWith(t, media.FixtureValid, t.TempDir())
	need := map[string]bool{
		"run_started":   false,
		"plan_pinned":   false,
		"worker_output": false,
		"verdict":       false,
		"run_accepted":  false,
	}
	for _, e := range res.Events {
		if _, ok := need[e.Kind]; ok {
			need[e.Kind] = true
		}
		if e.Kind == "run_started" && (e.ObjectiveID == "" || e.CriteriaHash == "") {
			t.Fatal("run_started missing objective/criteria")
		}
	}
	for k, v := range need {
		if !v {
			t.Fatalf("missing event kind %s", k)
		}
	}
	proj := kernel.Fold(res.Events)
	if proj.Status != res.Status {
		t.Fatalf("refold %s vs %s", proj.Status, res.Status)
	}
}

func TestHashChainDetectsTamper(t *testing.T) {
	dir := t.TempDir()
	_ = runWith(t, media.FixtureValid, dir)
	logPath := filepath.Join(dir, "provenance.jsonl")
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) < 2 {
		t.Fatal("expected multiple events")
	}
	lines[1] = lines[1][:len(lines[1])-1] + "X"
	if err := os.WriteFile(logPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := kernel.OpenLog(logPath); err == nil {
		t.Fatal("expected chain verification failure")
	}
}

func TestSameIdentityVerdictIgnoredByFold(t *testing.T) {
	events := []kernel.Event{
		{Kind: "run_started", ObjectiveID: "o", CriteriaHash: "c"},
		{Kind: "worker_output", Actor: media.ProducerID, ArtifactHash: "a1"},
		{Kind: "verdict", Actor: media.ProducerID, ArtifactHash: "a1", CriteriaHash: "c", Result: "pass"},
	}
	p := kernel.Fold(events)
	if p.Status == "accepted" {
		t.Fatal("same-identity pass must not accept")
	}
}

func TestMediaFieldsAbsentFromKernelSources(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"ProductionPackage", "narration", "audience", "Scenes", "FFmpeg", "timeline"}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		b, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, f := range forbidden {
			if strings.Contains(src, f) {
				t.Fatalf("kernel file %s contains media term %q", e.Name(), f)
			}
		}
	}
}

func TestBoundedRetriesExhaust(t *testing.T) {
	attempts := 0
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      t.TempDir(),
		Objective:    kernel.Objective{ID: "obj-bound", Text: "bounds"},
		CriteriaJSON: loadCriteria(t),
		Bounds:       kernel.Bounds{MaxAttempts: 2, MaxBytes: 1 << 20},
		Propose: func(obj kernel.Objective, planHash string) (kernel.Candidate, error) {
			attempts++
			return media.Propose(media.FixtureInvalid)(obj, planHash)
		},
		Transform:  media.Transform,
		Verify:     media.Verify,
		VerifierID: media.VerifierID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "rejected" {
		t.Fatalf("got %s", res.Status)
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d want 2", attempts)
	}
}

func TestZeroThirdPartyModules(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "github.com/smcdaniel54/Tamvori") {
			t.Fatalf("unexpected require: %s", line)
		}
	}
	if strings.Contains(string(b), "\nrequire (") {
		t.Fatal("require block present")
	}
}
