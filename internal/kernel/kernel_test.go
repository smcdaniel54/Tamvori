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
	// Walk up from this test file's module root.
	path := filepath.Join("..", "..", "experiments", "exp-001b-kernel-slice", "acceptance", "criteria.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read criteria: %v", err)
	}
	return b
}

func runWith(t *testing.T, mode media.FixtureMode, dir string) kernel.RunResult {
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

func TestValidRunAccepted(t *testing.T) {
	dir := t.TempDir()
	res := runWith(t, media.FixtureValid, dir)
	if res.Status != "accepted" {
		t.Fatalf("status=%s reason=%s", res.Status, res.RejectReason)
	}
	if res.ArtifactHash == "" {
		t.Fatal("missing artifact hash")
	}
	var sawVerdictPass bool
	for _, e := range res.Events {
		if e.Kind == kernel.EventVerdict && e.Result == "pass" {
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
		if e.Kind == kernel.EventClaim {
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
	pinned := kernel.CriteriaHashOf(criteria)
	dir := t.TempDir()
	acceptance := filepath.Join(dir, "acceptance")
	if err := os.MkdirAll(acceptance, 0o755); err != nil {
		t.Fatal(err)
	}
	critPath := filepath.Join(acceptance, "criteria.json")
	if err := os.WriteFile(critPath, criteria, 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate a worker attempting to write the acceptance directory.
	workerWrite := filepath.Join(acceptance, "hacked.json")
	if !kernel.UnderDir(acceptance, workerWrite) {
		t.Fatal("expected worker write path under acceptance dir")
	}
	// Policy: refuse. Criteria used for the run are the in-memory pinned bytes.
	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      filepath.Join(dir, "run"),
		Objective:    kernel.Objective{ID: "obj-pin", Text: "pin test"},
		CriteriaJSON: criteria, // pinned copy, not re-read from disk after mutation
		Bounds:       kernel.Bounds{MaxAttempts: 1, MaxBytes: 1 << 20},
		Propose: func(obj kernel.Objective, planHash string) (kernel.Candidate, error) {
			// Attempt to alter on-disk criteria during propose.
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
	if kernel.CriteriaHashOf(disk) == pinned {
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
	store := kernel.Store{Dir: t.TempDir()}
	h1, err := store.Put([]byte(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	h2, err := store.Put([]byte(`{"a":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Fatal("different bytes must different hashes")
	}
	h1b, err := store.Put([]byte(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h1b {
		t.Fatal("same bytes must same hash")
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
	var pkg media.ProductionPackage
	if err := json.Unmarshal(o1, &pkg); err != nil {
		t.Fatal(err)
	}
	if len(pkg.Scenes) < 2 || pkg.Scenes[0].ID > pkg.Scenes[1].ID {
		// sorted ascending
		for i := 1; i < len(pkg.Scenes); i++ {
			if pkg.Scenes[i-1].ID > pkg.Scenes[i].ID {
				t.Fatal("scenes not sorted")
			}
		}
	}
}

func TestProvenanceCapturesEvidence(t *testing.T) {
	res := runWith(t, media.FixtureValid, t.TempDir())
	need := map[string]bool{
		kernel.EventRunStarted:  false,
		kernel.EventPlanPinned:  false,
		kernel.EventWorkerOutput: false,
		kernel.EventVerdict:     false,
		kernel.EventRunAccepted: false,
	}
	for _, e := range res.Events {
		if _, ok := need[e.Kind]; ok {
			need[e.Kind] = true
		}
		if e.Kind == kernel.EventRunStarted && (e.ObjectiveID == "" || e.CriteriaHash == "") {
			t.Fatal("run_started missing objective/criteria")
		}
	}
	for k, v := range need {
		if !v {
			t.Fatalf("missing event kind %s", k)
		}
	}
	// Refold must match.
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
	// Tamper middle line.
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
		{Kind: kernel.EventRunStarted, ObjectiveID: "o", CriteriaHash: "c"},
		{Kind: kernel.EventWorkerOutput, Actor: media.ProducerID, ArtifactHash: "a1"},
		{Kind: kernel.EventVerdict, Actor: media.ProducerID, ArtifactHash: "a1", CriteriaHash: "c", Result: "pass"},
	}
	p := kernel.Fold(events)
	if p.Status == "accepted" {
		t.Fatal("same-identity pass must not accept")
	}
}

func TestMediaFieldsAbsentFromKernelSources(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		// tests run with cwd = package dir
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
