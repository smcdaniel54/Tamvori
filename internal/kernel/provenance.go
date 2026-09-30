package kernel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Event kinds for the append-only provenance log.
const (
	EventRunStarted       = "run_started"
	EventPlanPinned       = "plan_pinned"
	EventWorkerOutput     = "worker_output"
	EventClaim            = "claim"
	EventVerdict          = "verdict"
	EventRunAccepted      = "run_accepted"
	EventRunRejected      = "run_rejected"
	EventRetry            = "retry"
)

// Event is one append-only provenance record.
type Event struct {
	Seq           int               `json:"seq"`
	PrevHash      string            `json:"prev_hash"`
	EventHash     string            `json:"event_hash"`
	Kind          string            `json:"kind"`
	Actor         string            `json:"actor"`
	ObjectiveID   string            `json:"objective_id,omitempty"`
	CriteriaHash  string            `json:"criteria_hash,omitempty"`
	PlanHash      string            `json:"plan_hash,omitempty"`
	ArtifactHash  string            `json:"artifact_hash,omitempty"`
	Result        string            `json:"result,omitempty"` // pass|fail for verdicts
	EvidenceHash  string            `json:"evidence_hash,omitempty"`
	Attempt       int               `json:"attempt,omitempty"`
	Note          string            `json:"note,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// eventBody is the portion hashed into EventHash (excludes EventHash itself).
type eventBody struct {
	Seq          int               `json:"seq"`
	PrevHash     string            `json:"prev_hash"`
	Kind         string            `json:"kind"`
	Actor        string            `json:"actor"`
	ObjectiveID  string            `json:"objective_id,omitempty"`
	CriteriaHash string            `json:"criteria_hash,omitempty"`
	PlanHash     string            `json:"plan_hash,omitempty"`
	ArtifactHash string            `json:"artifact_hash,omitempty"`
	Result       string            `json:"result,omitempty"`
	EvidenceHash string            `json:"evidence_hash,omitempty"`
	Attempt      int               `json:"attempt,omitempty"`
	Note         string            `json:"note,omitempty"`
	Extra        map[string]string `json:"extra,omitempty"`
}

func (e Event) bodyHash() (string, error) {
	b := eventBody{
		Seq: e.Seq, PrevHash: e.PrevHash, Kind: e.Kind, Actor: e.Actor,
		ObjectiveID: e.ObjectiveID, CriteriaHash: e.CriteriaHash, PlanHash: e.PlanHash,
		ArtifactHash: e.ArtifactHash, Result: e.Result, EvidenceHash: e.EvidenceHash,
		Attempt: e.Attempt, Note: e.Note, Extra: e.Extra,
	}
	raw, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	return HashBytes(raw), nil
}

// Log is an append-only hash-chained event file.
type Log struct {
	Path   string
	events []Event
}

func OpenLog(path string) (*Log, error) {
	l := &Log{Path: path}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	// Increase buffer for long lines.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	var prev string
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("log parse: %w", err)
		}
		want, err := e.bodyHash()
		if err != nil {
			return nil, err
		}
		if e.EventHash != want {
			return nil, fmt.Errorf("log chain broken at seq %d: event hash mismatch", e.Seq)
		}
		if e.PrevHash != prev {
			return nil, fmt.Errorf("log chain broken at seq %d: prev hash mismatch", e.Seq)
		}
		prev = e.EventHash
		l.events = append(l.events, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Log) Events() []Event {
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

func (l *Log) Append(e Event) (Event, error) {
	e.Seq = len(l.events) + 1
	if len(l.events) == 0 {
		e.PrevHash = ""
	} else {
		e.PrevHash = l.events[len(l.events)-1].EventHash
	}
	h, err := e.bodyHash()
	if err != nil {
		return Event{}, err
	}
	e.EventHash = h
	raw, err := json.Marshal(e)
	if err != nil {
		return Event{}, err
	}
	f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return Event{}, err
	}
	defer f.Close()
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return Event{}, err
	}
	l.events = append(l.events, e)
	return e, nil
}

// Projection is the pure fold of a log into run status.
type Projection struct {
	Status       string // drafted|executing|accepted|rejected
	ObjectiveID  string
	CriteriaHash string
	PlanHash     string
	ArtifactHash string
	LastVerdict  string
	LastActor    string
}

// Fold recomputes status from events. Pure: no I/O.
// Claims never move status to accepted. Verdicts from the same actor as the
// producer of the artifact under review are ignored for acceptance.
func Fold(events []Event) Projection {
	p := Projection{Status: "drafted"}
	producerOf := map[string]string{} // artifact hash -> actor
	for _, e := range events {
		switch e.Kind {
		case EventRunStarted:
			p.ObjectiveID = e.ObjectiveID
			p.CriteriaHash = e.CriteriaHash
			p.Status = "executing"
		case EventPlanPinned:
			p.PlanHash = e.PlanHash
		case EventWorkerOutput:
			p.ArtifactHash = e.ArtifactHash
			producerOf[e.ArtifactHash] = e.Actor
		case EventClaim:
			// Ignored for status.
		case EventVerdict:
			p.LastVerdict = e.Result
			p.LastActor = e.Actor
			if e.ArtifactHash != "" {
				if prod, ok := producerOf[e.ArtifactHash]; ok && prod == e.Actor {
					// Same identity: ignore for acceptance.
					continue
				}
			}
			if e.CriteriaHash != "" && p.CriteriaHash != "" && e.CriteriaHash != p.CriteriaHash {
				continue
			}
			if e.Result == "pass" {
				p.Status = "accepted"
				p.ArtifactHash = e.ArtifactHash
			} else if e.Result == "fail" {
				p.Status = "rejected"
			}
		case EventRunAccepted:
			p.Status = "accepted"
			if e.ArtifactHash != "" {
				p.ArtifactHash = e.ArtifactHash
			}
		case EventRunRejected:
			p.Status = "rejected"
		}
	}
	return p
}
