package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
)

// SchemaID identifies the media production package adapter schema.
const SchemaID = "tamvori.media.production_package.v1"

// Scene is one outline entry in the production package.
type Scene struct {
	ID      string `json:"id"`
	Outline string `json:"outline"`
}

// Asset is a named content-addressed reference (no media bytes).
type Asset struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

// ProductionPackage is the structured media workload artifact.
// It lives only in this adapter package. The kernel never imports these fields into Fold.
type ProductionPackage struct {
	Schema    string  `json:"schema"`
	Title     string  `json:"title"`
	Audience  string  `json:"audience"`
	Message   string  `json:"message"`
	Scenes    []Scene `json:"scenes"`
	Narration string  `json:"narration"`
	Assets    []Asset `json:"assets"`
}

// Criteria is the pinned acceptance document for this adapter (JSON in acceptance/).
type Criteria struct {
	Schema              string   `json:"schema"`
	RequiredFields      []string `json:"required_fields"`
	RequireSortedScenes bool     `json:"require_sorted_scenes"`
	RequireSortedAssets bool     `json:"require_sorted_assets"`
	RequireAssetHashes  bool     `json:"require_asset_hashes"`
	MinScenes           int      `json:"min_scenes"`
	ForbiddenTitles     []string `json:"forbidden_titles"`
}

// ProducerID is the generative-shaped fixture identity.
const ProducerID = "worker/fixture-gen"

// VerifierID is the independent media verifier identity.
const VerifierID = "verifier/media-package"

// FixtureMode selects what the generative-shaped worker emits.
type FixtureMode string

const (
	FixtureValid     FixtureMode = "valid"
	FixtureInvalid   FixtureMode = "invalid"
	FixtureClaimPass FixtureMode = "claim_pass" // invalid bytes + claim valid:true
)

// Propose returns a generative-shaped candidate. It does not see criteria bytes.
func Propose(mode FixtureMode) kernel.ProposeFunc {
	return func(obj kernel.Objective, planHash string) (kernel.Candidate, error) {
		_ = planHash
		switch mode {
		case FixtureValid, "":
			raw, err := json.Marshal(draftValid(obj))
			if err != nil {
				return kernel.Candidate{}, err
			}
			return kernel.Candidate{Bytes: raw, Producer: ProducerID}, nil
		case FixtureInvalid:
			// Missing required fields; also claims success (must be ignored).
			raw := []byte(`{"schema":"tamvori.media.production_package.v1","title":"","claim":true}`)
			return kernel.Candidate{
				Bytes: raw, Producer: ProducerID,
				ClaimValid: true, ClaimNote: "fixture asserts valid",
			}, nil
		case FixtureClaimPass:
			raw, err := json.Marshal(map[string]any{
				"schema": SchemaID,
				"title":  "broken",
				// missing audience, message, scenes, narration, assets
			})
			if err != nil {
				return kernel.Candidate{}, err
			}
			return kernel.Candidate{
				Bytes: raw, Producer: ProducerID,
				ClaimValid: true, ClaimNote: "I verified myself",
			}, nil
		default:
			return kernel.Candidate{}, fmt.Errorf("unknown fixture mode %q", mode)
		}
	}
}

func draftValid(obj kernel.Objective) ProductionPackage {
	title := "Demo package"
	if obj.Text != "" {
		title = "Package for: " + obj.Text
	}
	return ProductionPackage{
		Schema:   SchemaID,
		Title:    title,
		Audience: "operators",
		Message:  "prove governed production",
		Scenes: []Scene{
			{ID: "2", Outline: "close"},
			{ID: "1", Outline: "open"},
		},
		Narration: "A short narration fragment.",
		Assets: []Asset{
			{Name: "b-roll", Hash: ""}, // filled by transform
			{Name: "a-roll", Hash: ""},
		},
	}
}

// Transform canonicalizes a candidate into a ProductionPackage.
// Same input bytes always produce the same output bytes.
func Transform(in []byte) ([]byte, error) {
	var loose map[string]json.RawMessage
	if err := json.Unmarshal(in, &loose); err != nil {
		return nil, fmt.Errorf("candidate is not JSON: %w", err)
	}
	var pkg ProductionPackage
	if err := json.Unmarshal(in, &pkg); err != nil {
		return nil, fmt.Errorf("candidate schema decode: %w", err)
	}
	if pkg.Schema == "" {
		pkg.Schema = SchemaID
	}
	// Deterministic ordering.
	sort.Slice(pkg.Scenes, func(i, j int) bool { return pkg.Scenes[i].ID < pkg.Scenes[j].ID })
	sort.Slice(pkg.Assets, func(i, j int) bool { return pkg.Assets[i].Name < pkg.Assets[j].Name })
	// Derive asset hashes from name + package title (deterministic, no files).
	for i := range pkg.Assets {
		pkg.Assets[i].Hash = kernel.HashBytes([]byte(pkg.Title + "|" + pkg.Assets[i].Name))
	}
	return marshalCanonical(pkg)
}

func marshalCanonical(pkg ProductionPackage) ([]byte, error) {
	// encoding/json struct order is stable for this type.
	b, err := json.Marshal(pkg)
	if err != nil {
		return nil, err
	}
	// Compact to a single canonical form (Marshal already compact).
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Verify checks the transformed artifact against pinned criteria JSON.
func Verify(artifact []byte, criteriaJSON []byte, criteriaHash string) (kernel.Verdict, error) {
	v := kernel.Verdict{
		Evaluator:    VerifierID,
		CriteriaHash: criteriaHash,
		ArtifactHash: kernel.HashBytes(artifact),
	}
	var crit Criteria
	if err := json.Unmarshal(criteriaJSON, &crit); err != nil {
		v.Pass = false
		v.Note = "criteria decode failed"
		return v, nil
	}
	var pkg ProductionPackage
	if err := json.Unmarshal(artifact, &pkg); err != nil {
		v.Pass = false
		v.Note = "artifact is not a production package"
		evidence, _ := json.Marshal(map[string]string{"error": err.Error()})
		v.Evidence = evidence
		return v, nil
	}

	var failures []string
	if pkg.Schema != SchemaID {
		failures = append(failures, "schema mismatch")
	}
	for _, f := range crit.RequiredFields {
		switch f {
		case "title":
			if strings.TrimSpace(pkg.Title) == "" {
				failures = append(failures, "missing title")
			}
		case "audience":
			if strings.TrimSpace(pkg.Audience) == "" {
				failures = append(failures, "missing audience")
			}
		case "message":
			if strings.TrimSpace(pkg.Message) == "" {
				failures = append(failures, "missing message")
			}
		case "scenes":
			if len(pkg.Scenes) == 0 {
				failures = append(failures, "missing scenes")
			}
		case "narration":
			if strings.TrimSpace(pkg.Narration) == "" {
				failures = append(failures, "missing narration")
			}
		case "assets":
			if len(pkg.Assets) == 0 {
				failures = append(failures, "missing assets")
			}
		}
	}
	if crit.MinScenes > 0 && len(pkg.Scenes) < crit.MinScenes {
		failures = append(failures, "too few scenes")
	}
	if crit.RequireSortedScenes && !scenesSorted(pkg.Scenes) {
		failures = append(failures, "scenes not sorted")
	}
	if crit.RequireSortedAssets && !assetsSorted(pkg.Assets) {
		failures = append(failures, "assets not sorted")
	}
	if crit.RequireAssetHashes {
		for _, a := range pkg.Assets {
			if a.Hash == "" || len(a.Hash) != 64 {
				failures = append(failures, "asset hash missing")
				break
			}
		}
	}
	for _, bad := range crit.ForbiddenTitles {
		if pkg.Title == bad {
			failures = append(failures, "forbidden title")
		}
	}

	evidence, _ := json.Marshal(map[string]any{
		"failures":      failures,
		"schema":        pkg.Schema,
		"criteria_hash": criteriaHash,
	})
	v.Evidence = evidence
	if len(failures) > 0 {
		v.Pass = false
		v.Note = strings.Join(failures, "; ")
		return v, nil
	}
	v.Pass = true
	v.Note = "criteria satisfied"
	return v, nil
}

func scenesSorted(s []Scene) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1].ID > s[i].ID {
			return false
		}
	}
	return true
}

func assetsSorted(a []Asset) bool {
	for i := 1; i < len(a); i++ {
		if a[i-1].Name > a[i].Name {
			return false
		}
	}
	return true
}
