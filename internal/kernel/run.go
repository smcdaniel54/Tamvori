package kernel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Objective is what a human wants. Criteria are pinned by hash at run start.
type Objective struct {
	ID   string
	Text string
}

// Bounds are the only "policy" for this slice: immutable run data.
type Bounds struct {
	MaxAttempts int
	MaxBytes    int
}

// Candidate is worker output: bytes plus optional self-claim. Claims have no authority.
type Candidate struct {
	Bytes      []byte
	Producer   string
	ClaimValid bool
	ClaimNote  string
}

// Verdict is an independent evaluation result.
type Verdict struct {
	Evaluator    string
	ArtifactHash string
	CriteriaHash string
	Pass         bool
	Evidence     []byte
	Note         string
}

// RunRequest is everything required to execute one governed run.
type RunRequest struct {
	WorkDir      string
	Objective    Objective
	CriteriaJSON []byte
	Bounds       Bounds
	Propose      func(Objective, string) (Candidate, error)
	Transform    func([]byte) ([]byte, error)
	Verify       func(artifact, criteria []byte, criteriaHash string) (Verdict, error)
	VerifierID   string
}

// RunResult is the outcome of one governed run.
type RunResult struct {
	Status        string
	CriteriaHash  string
	PlanHash      string
	CandidateHash string
	ArtifactHash  string
	Projection    Projection
	Events        []Event
	RejectReason  string
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

	st := store{Dir: filepath.Join(req.WorkDir, "artifacts")}
	log, err := OpenLog(filepath.Join(req.WorkDir, "provenance.jsonl"))
	if err != nil {
		return RunResult{}, err
	}

	criteriaHash := HashBytes(req.CriteriaJSON)
	if _, err := st.put(req.CriteriaJSON); err != nil {
		return RunResult{}, err
	}

	if _, err := log.append(Event{
		Kind: kindRunStarted, Actor: "governor",
		ObjectiveID: req.Objective.ID, CriteriaHash: criteriaHash,
		Note: req.Objective.Text,
	}); err != nil {
		return RunResult{}, err
	}

	planBytes, err := json.Marshal(map[string][]string{
		"steps": {"propose", "transform", "verify"},
	})
	if err != nil {
		return RunResult{}, err
	}
	planHash, err := st.put(planBytes)
	if err != nil {
		return RunResult{}, err
	}
	if _, err := log.append(Event{
		Kind: kindPlanPinned, Actor: "governor",
		CriteriaHash: criteriaHash, PlanHash: planHash,
	}); err != nil {
		return RunResult{}, err
	}

	var lastReject, lastCandidateHash, lastArtifactHash string

	for attempt := 1; attempt <= req.Bounds.MaxAttempts; attempt++ {
		if attempt > 1 {
			if _, err := log.append(Event{
				Kind: kindRetry, Actor: "governor",
				CriteriaHash: criteriaHash, Attempt: attempt, Note: lastReject,
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

		candHash, err := st.put(cand.Bytes)
		if err != nil {
			return RunResult{}, err
		}
		lastCandidateHash = candHash
		if _, err := log.append(Event{
			Kind: kindWorkerOutput, Actor: cand.Producer,
			CriteriaHash: criteriaHash, ArtifactHash: candHash, Attempt: attempt,
			Note: "propose",
		}); err != nil {
			return RunResult{}, err
		}

		if cand.ClaimValid || cand.ClaimNote != "" {
			claimBytes, _ := json.Marshal(map[string]any{"valid": cand.ClaimValid, "note": cand.ClaimNote})
			claimHash, err := st.put(claimBytes)
			if err != nil {
				return RunResult{}, err
			}
			if _, err := log.append(Event{
				Kind: kindClaim, Actor: cand.Producer,
				CriteriaHash: criteriaHash, ArtifactHash: candHash,
				EvidenceHash: claimHash, Note: "worker self-claim",
			}); err != nil {
				return RunResult{}, err
			}
		}

		out, err := req.Transform(cand.Bytes)
		if err != nil {
			lastReject = "transform: " + err.Error()
			if _, err := log.append(Event{
				Kind: kindVerdict, Actor: req.VerifierID,
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
		artHash, err := st.put(out)
		if err != nil {
			return RunResult{}, err
		}
		lastArtifactHash = artHash
		if _, err := log.append(Event{
			Kind: kindWorkerOutput, Actor: "worker/deterministic",
			CriteriaHash: criteriaHash, ArtifactHash: artHash, Attempt: attempt,
			Note: "transform input=" + candHash,
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
			lastReject = "verifier identity equals producer"
			if _, err := log.append(Event{
				Kind: kindVerdict, Actor: v.Evaluator,
				CriteriaHash: criteriaHash, ArtifactHash: artHash,
				Result: "fail", Attempt: attempt, Note: lastReject,
			}); err != nil {
				return RunResult{}, err
			}
			continue
		}

		var evidenceHash string
		if len(v.Evidence) > 0 {
			evidenceHash, err = st.put(v.Evidence)
			if err != nil {
				return RunResult{}, err
			}
		}
		result := "fail"
		if v.Pass {
			result = "pass"
		}
		if _, err := log.append(Event{
			Kind: kindVerdict, Actor: v.Evaluator,
			CriteriaHash: criteriaHash, ArtifactHash: artHash,
			Result: result, EvidenceHash: evidenceHash, Attempt: attempt, Note: v.Note,
		}); err != nil {
			return RunResult{}, err
		}

		if v.Pass {
			if _, err := log.append(Event{
				Kind: kindRunAccepted, Actor: "governor",
				CriteriaHash: criteriaHash, ArtifactHash: artHash,
			}); err != nil {
				return RunResult{}, err
			}
			return finish(log, criteriaHash, planHash, lastCandidateHash, artHash, "")
		}
		lastReject = v.Note
		if lastReject == "" {
			lastReject = "verification failed"
		}
	}

	if _, err := log.append(Event{
		Kind: kindRunRejected, Actor: "governor",
		CriteriaHash: criteriaHash, ArtifactHash: lastArtifactHash,
		Note: lastReject,
	}); err != nil {
		return RunResult{}, err
	}
	return finish(log, criteriaHash, planHash, lastCandidateHash, lastArtifactHash, lastReject)
}

func finish(log *Log, criteriaHash, planHash, candidateHash, artifactHash, reject string) (RunResult, error) {
	proj := Fold(log.Events())
	return RunResult{
		Status: proj.Status, CriteriaHash: criteriaHash, PlanHash: planHash,
		CandidateHash: candidateHash, ArtifactHash: artifactHash,
		Projection: proj, Events: log.Events(), RejectReason: reject,
	}, nil
}
