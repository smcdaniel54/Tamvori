package kernel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Objective is what a human wants. Criteria are pinned by hash at run start.
type Objective struct {
	ID   string
	Text string
}

// Bounds are the only "policy" for this slice: immutable run data, not a framework.
type Bounds struct {
	MaxAttempts int
	MaxBytes    int
}

// Plan is ordered step names as data.
type Plan struct {
	Steps []string `json:"steps"`
}

// Candidate is worker output: bytes plus optional self-claim. Claims have no authority.
type Candidate struct {
	Bytes       []byte
	Producer    string
	ClaimValid  bool // worker self-report; ignored by Fold for acceptance
	ClaimNote   string
}

// Verdict is an independent evaluation result.
type Verdict struct {
	Evaluator    string
	ArtifactHash string
	CriteriaHash string
	Pass         bool
	Evidence     []byte // optional evidence bytes stored as artifact
	Note         string
}

// VerifyFunc evaluates an artifact against pinned criteria bytes.
// It must not be the producer identity.
type VerifyFunc func(artifact []byte, criteria []byte, criteriaHash string) (Verdict, error)

// TransformFunc is a deterministic worker: same input bytes -> same output bytes.
type TransformFunc func(in []byte) ([]byte, error)

// ProposeFunc is a generative-shaped worker (fixture in this experiment).
type ProposeFunc func(objective Objective, planHash string) (Candidate, error)

// RunRequest is everything required to execute one governed run.
type RunRequest struct {
	WorkDir      string
	Objective    Objective
	CriteriaJSON []byte // pinned; workers do not receive write access
	Bounds       Bounds
	Propose      ProposeFunc
	Transform    TransformFunc
	Verify       VerifyFunc
	VerifierID   string // must differ from producer
}

// RunResult is the outcome of one governed run.
type RunResult struct {
	Status         string
	CriteriaHash   string
	PlanHash       string
	CandidateHash  string
	ArtifactHash   string
	Projection     Projection
	Events         []Event
	RejectReason   string
}

// Execute runs one governed objective-to-artifact cycle.
func Execute(req RunRequest) (RunResult, error) {
	if req.Bounds.MaxAttempts <= 0 {
		req.Bounds.MaxAttempts = 1
	}
	if req.Bounds.MaxBytes <= 0 {
		req.Bounds.MaxBytes = 1 << 20
	}
	if err := os.MkdirAll(req.WorkDir, 0o755); err != nil {
		return RunResult{}, err
	}

	store := Store{Dir: filepath.Join(req.WorkDir, "artifacts")}
	logPath := filepath.Join(req.WorkDir, "provenance.jsonl")
	log, err := OpenLog(logPath)
	if err != nil {
		return RunResult{}, err
	}

	criteriaHash := HashBytes(req.CriteriaJSON)
	if _, err := store.Put(req.CriteriaJSON); err != nil {
		return RunResult{}, err
	}

	if _, err := log.Append(Event{
		Kind: EventRunStarted, Actor: "governor",
		ObjectiveID: req.Objective.ID, CriteriaHash: criteriaHash,
		Note: req.Objective.Text,
	}); err != nil {
		return RunResult{}, err
	}

	plan := Plan{Steps: []string{"propose", "transform", "verify"}}
	planBytes, err := json.Marshal(plan)
	if err != nil {
		return RunResult{}, err
	}
	planHash, err := store.Put(planBytes)
	if err != nil {
		return RunResult{}, err
	}
	if _, err := log.Append(Event{
		Kind: EventPlanPinned, Actor: "governor",
		CriteriaHash: criteriaHash, PlanHash: planHash,
	}); err != nil {
		return RunResult{}, err
	}

	var lastReject string
	var lastCandidateHash string
	var lastArtifactHash string

	for attempt := 1; attempt <= req.Bounds.MaxAttempts; attempt++ {
		if attempt > 1 {
			if _, err := log.Append(Event{
				Kind: EventRetry, Actor: "governor",
				CriteriaHash: criteriaHash, Attempt: attempt,
				Note: lastReject,
			}); err != nil {
				return RunResult{}, err
			}
		}

		cand, err := req.Propose(req.Objective, planHash)
		if err != nil {
			return RunResult{}, err
		}
		if cand.Producer == "" {
			return RunResult{}, fmt.Errorf("propose: missing producer identity")
		}
		if len(cand.Bytes) > req.Bounds.MaxBytes {
			lastReject = "candidate exceeds MaxBytes"
			continue
		}

		candHash, err := store.Put(cand.Bytes)
		if err != nil {
			return RunResult{}, err
		}
		lastCandidateHash = candHash
		if _, err := log.Append(Event{
			Kind: EventWorkerOutput, Actor: cand.Producer,
			CriteriaHash: criteriaHash, ArtifactHash: candHash, Attempt: attempt,
			Note: "propose",
		}); err != nil {
			return RunResult{}, err
		}

		if cand.ClaimValid || cand.ClaimNote != "" {
			claim := map[string]any{"valid": cand.ClaimValid, "note": cand.ClaimNote}
			claimBytes, _ := json.Marshal(claim)
			claimHash, err := store.Put(claimBytes)
			if err != nil {
				return RunResult{}, err
			}
			if _, err := log.Append(Event{
				Kind: EventClaim, Actor: cand.Producer,
				CriteriaHash: criteriaHash, ArtifactHash: candHash,
				EvidenceHash: claimHash, Note: "worker self-claim",
			}); err != nil {
				return RunResult{}, err
			}
		}

		out, err := req.Transform(cand.Bytes)
		if err != nil {
			lastReject = "transform: " + err.Error()
			if _, err := log.Append(Event{
				Kind: EventVerdict, Actor: req.VerifierID,
				CriteriaHash: criteriaHash, ArtifactHash: candHash,
				Result: "fail", Attempt: attempt, Note: lastReject,
			}); err != nil {
				return RunResult{}, err
			}
			continue
		}
		if len(out) > req.Bounds.MaxBytes {
			lastReject = "transformed artifact exceeds MaxBytes"
			continue
		}
		artHash, err := store.Put(out)
		if err != nil {
			return RunResult{}, err
		}
		lastArtifactHash = artHash
		if _, err := log.Append(Event{
			Kind: EventWorkerOutput, Actor: "worker/deterministic",
			CriteriaHash: criteriaHash, ArtifactHash: artHash, Attempt: attempt,
			Note: "transform", Extra: map[string]string{"input": candHash},
		}); err != nil {
			return RunResult{}, err
		}

		v, err := req.Verify(out, req.CriteriaJSON, criteriaHash)
		if err != nil {
			return RunResult{}, err
		}
		if v.Evaluator == "" {
			v.Evaluator = req.VerifierID
		}
		if v.Evaluator == cand.Producer {
			// Hard rule: refuse to accept a same-identity verdict as authority.
			// Record it; Fold will ignore it for acceptance. Treat as fail path.
			lastReject = "verifier identity equals producer"
			if _, err := log.Append(Event{
				Kind: EventVerdict, Actor: v.Evaluator,
				CriteriaHash: criteriaHash, ArtifactHash: artHash,
				Result: "fail", Attempt: attempt, Note: lastReject,
			}); err != nil {
				return RunResult{}, err
			}
			continue
		}

		var evidenceHash string
		if len(v.Evidence) > 0 {
			evidenceHash, err = store.Put(v.Evidence)
			if err != nil {
				return RunResult{}, err
			}
		}
		result := "fail"
		if v.Pass {
			result = "pass"
		}
		if _, err := log.Append(Event{
			Kind: EventVerdict, Actor: v.Evaluator,
			CriteriaHash: criteriaHash, ArtifactHash: artHash,
			Result: result, EvidenceHash: evidenceHash, Attempt: attempt, Note: v.Note,
		}); err != nil {
			return RunResult{}, err
		}

		if v.Pass {
			if _, err := log.Append(Event{
				Kind: EventRunAccepted, Actor: "governor",
				CriteriaHash: criteriaHash, ArtifactHash: artHash,
			}); err != nil {
				return RunResult{}, err
			}
			proj := Fold(log.Events())
			return RunResult{
				Status: proj.Status, CriteriaHash: criteriaHash, PlanHash: planHash,
				CandidateHash: lastCandidateHash, ArtifactHash: artHash,
				Projection: proj, Events: log.Events(),
			}, nil
		}
		lastReject = v.Note
		if lastReject == "" {
			lastReject = "verification failed"
		}
	}

	if _, err := log.Append(Event{
		Kind: EventRunRejected, Actor: "governor",
		CriteriaHash: criteriaHash, ArtifactHash: lastArtifactHash,
		Note: lastReject,
	}); err != nil {
		return RunResult{}, err
	}
	proj := Fold(log.Events())
	return RunResult{
		Status: proj.Status, CriteriaHash: criteriaHash, PlanHash: planHash,
		CandidateHash: lastCandidateHash, ArtifactHash: lastArtifactHash,
		Projection: proj, Events: log.Events(), RejectReason: lastReject,
	}, nil
}

// CriteriaHashOf returns HashBytes(criteria) for tests and CLI.
func CriteriaHashOf(criteria []byte) string { return HashBytes(criteria) }

// UnderDir reports whether writePath is inside dir (workers must not write criteria dirs).
func UnderDir(dir, writePath string) bool {
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
