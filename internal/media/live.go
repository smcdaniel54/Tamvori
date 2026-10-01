package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/openai"
)

// LiveProducerID is the generative OpenAI worker identity.
const LiveProducerID = "worker/openai-propose"

// DefaultLiveModel is a small inexpensive model for EXP-001C.
const DefaultLiveModel = "gpt-4o-mini"

const liveSystem = `You draft a Tamvori media production package candidate as JSON only.
Return one JSON object with exactly these keys:
schema, title, audience, message, scenes, narration, assets.
schema must be "tamvori.media.production_package.v1".
Hard requirements (deterministic verifier will reject otherwise):
- title, audience, message, and narration must be non-empty after trimming whitespace
- scenes must contain at least one object; each scene id and outline must be non-empty after trim
- assets must contain at least one object; each asset name must be a non-empty meaningful identifier after trim (for example "a-roll", not "")
- asset hash may be an empty string; do not invent verification claims
Do not include valid/passed/approved/score/confidence fields.
Do not claim the package is accepted.`

// LiveConfig configures one concrete OpenAI propose path. No provider registry.
type LiveConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
	// Context may carry test endpoint/client overrides from openai.With*.
	Context context.Context
}

// ProposeLive returns a ProposeFunc that performs one real (or test-injected) OpenAI call.
// Credentials come from LiveConfig / environment — never from criteria or git.
func ProposeLive(cfg LiveConfig) func(kernel.Objective, string) (kernel.Candidate, error) {
	return func(obj kernel.Objective, planHash string) (kernel.Candidate, error) {
		key := cfg.APIKey
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		model := cfg.Model
		if model == "" {
			model = DefaultLiveModel
		}
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = openai.DefaultTimeout
		}
		parent := cfg.Context
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		user := fmt.Sprintf(
			"Objective ID: %s\nObjective: %s\nPlan hash (context only): %s\nProduce the JSON package candidate now.",
			obj.ID, obj.Text, planHash,
		)
		content, raw, err := openai.ChatJSON(ctx, key, model, liveSystem, user, 800)
		if err != nil {
			c := kernel.Candidate{Producer: LiveProducerID}
			if len(raw) > 0 {
				c.Evidence = raw
				c.EvidenceNote = evidenceNote(model, raw)
			}
			return c, fmt.Errorf("live propose: %w", err)
		}
		content = strings.TrimSpace(content)
		if !json.Valid([]byte(content)) {
			return kernel.Candidate{
				Bytes:        []byte(content),
				Producer:     LiveProducerID,
				Evidence:     raw,
				EvidenceNote: evidenceNote(model, raw),
			}, fmt.Errorf("live propose: assistant content is not JSON")
		}
		// Strip any authority-claim fields the model may have invented.
		sanitized, err := stripAuthorityFields([]byte(content))
		if err != nil {
			return kernel.Candidate{}, err
		}
		return kernel.Candidate{
			Bytes:        sanitized,
			Producer:     LiveProducerID,
			Evidence:     raw,
			EvidenceNote: evidenceNote(model, raw),
		}, nil
	}
}

// ProposeFrozen replays a previously captured assistant JSON content blob.
// Used for deterministic offline replay of a live response. No network.
// Bytes are used as captured (already past live sanitization).
func ProposeFrozen(assistantJSON []byte, model string, rawEvidence []byte) func(kernel.Objective, string) (kernel.Candidate, error) {
	return func(kernel.Objective, string) (kernel.Candidate, error) {
		if !json.Valid(assistantJSON) {
			return kernel.Candidate{}, fmt.Errorf("frozen propose: not JSON")
		}
		ev := rawEvidence
		if len(ev) == 0 {
			ev = assistantJSON
		}
		if model == "" {
			model = DefaultLiveModel
		}
		return kernel.Candidate{
			Bytes:        append([]byte(nil), assistantJSON...),
			Producer:     LiveProducerID,
			Evidence:     ev,
			EvidenceNote: evidenceNote(model, ev),
		}, nil
	}
}

func evidenceNote(model string, raw []byte) string {
	return fmt.Sprintf("provider=openai model=%s response_hash=%s", model, kernel.HashBytes(raw))
}

func stripAuthorityFields(b []byte) ([]byte, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("live propose: parse candidate: %w", err)
	}
	for _, k := range []string{"valid", "passed", "approved", "score", "confidence"} {
		delete(m, k)
	}
	out, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return out, nil
}
