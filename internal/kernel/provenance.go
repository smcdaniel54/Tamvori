package kernel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Event kinds (private). External checks compare Event.Kind string values.
const (
	kindRunStarted   = "run_started"
	kindPlanPinned   = "plan_pinned"
	kindWorkerOutput = "worker_output"
	kindClaim        = "claim"
	kindVerdict      = "verdict"
	kindRunAccepted  = "run_accepted"
	kindRunRejected  = "run_rejected"
	kindRetry        = "retry"
)

// Event is one append-only provenance record.
type Event struct {
	Seq          int    `json:"seq"`
	PrevHash     string `json:"prev_hash"`
	EventHash    string `json:"event_hash,omitempty"`
	Kind         string `json:"kind"`
	Actor        string `json:"actor"`
	ObjectiveID  string `json:"objective_id,omitempty"`
	CriteriaHash string `json:"criteria_hash,omitempty"`
	PlanHash     string `json:"plan_hash,omitempty"`
	ArtifactHash string `json:"artifact_hash,omitempty"`
	Result       string `json:"result,omitempty"` // pass|fail for verdicts
	EvidenceHash string `json:"evidence_hash,omitempty"`
	Attempt      int    `json:"attempt,omitempty"`
	Note         string `json:"note,omitempty"`
}

func (e Event) bodyHash() (string, error) {
	cp := e
	cp.EventHash = ""
	raw, err := json.Marshal(cp)
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

// OpenLog reads and verifies an existing provenance log, or returns empty.
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

// Events returns a copy of the log.
func (l *Log) Events() []Event {
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

func (l *Log) append(e Event) (Event, error) {
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
}

// Fold recomputes status from events. Pure: no I/O.
// Claims never move status to accepted. Verdicts from the same actor as the
// producer of the artifact under review are ignored for acceptance.
func Fold(events []Event) Projection {
	p := Projection{Status: "drafted"}
	producerOf := map[string]string{} // artifact hash -> actor
	for _, e := range events {
		switch e.Kind {
		case kindRunStarted:
			p.ObjectiveID = e.ObjectiveID
			p.CriteriaHash = e.CriteriaHash
			p.Status = "executing"
		case kindPlanPinned:
			p.PlanHash = e.PlanHash
		case kindWorkerOutput:
			p.ArtifactHash = e.ArtifactHash
			producerOf[e.ArtifactHash] = e.Actor
		case kindClaim:
			// Ignored for status.
		case kindVerdict:
			if e.ArtifactHash != "" {
				if prod, ok := producerOf[e.ArtifactHash]; ok && prod == e.Actor {
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
		case kindRunAccepted:
			p.Status = "accepted"
			if e.ArtifactHash != "" {
				p.ArtifactHash = e.ArtifactHash
			}
		case kindRunRejected:
			p.Status = "rejected"
		}
	}
	return p
}
