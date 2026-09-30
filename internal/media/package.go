package media

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
)

const schemaID = "tamvori.media.production_package.v1"

// ProducerID is the generative-shaped fixture identity.
const ProducerID = "worker/fixture-gen"

// VerifierID is the independent media verifier identity.
const VerifierID = "verifier/media-package"

// Fixture modes for the generative-shaped worker.
const (
	FixtureValid     = "valid"
	FixtureInvalid   = "invalid"
	FixtureClaimPass = "claim_pass"
)

type scene struct {
	ID      string `json:"id"`
	Outline string `json:"outline"`
}

type asset struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type productionPackage struct {
	Schema    string  `json:"schema"`
	Title     string  `json:"title"`
	Audience  string  `json:"audience"`
	Message   string  `json:"message"`
	Scenes    []scene `json:"scenes"`
	Narration string  `json:"narration"`
	Assets    []asset `json:"assets"`
}

type criteria struct {
	RequiredFields      []string `json:"required_fields"`
	RequireSortedScenes bool     `json:"require_sorted_scenes"`
	RequireSortedAssets bool     `json:"require_sorted_assets"`
	RequireAssetHashes  bool     `json:"require_asset_hashes"`
	MinScenes           int      `json:"min_scenes"`
	ForbiddenTitles     []string `json:"forbidden_titles"`
}

// Propose returns a generative-shaped candidate. It does not see criteria bytes.
func Propose(mode string) func(kernel.Objective, string) (kernel.Candidate, error) {
	return func(obj kernel.Objective, _ string) (kernel.Candidate, error) {
		switch mode {
		case FixtureValid, "":
			raw, err := json.Marshal(draftValid(obj))
			if err != nil {
				return kernel.Candidate{}, err
			}
			return kernel.Candidate{Bytes: raw, Producer: ProducerID}, nil
		case FixtureInvalid:
			raw := []byte(`{"schema":"tamvori.media.production_package.v1","title":"","claim":true}`)
			return kernel.Candidate{
				Bytes: raw, Producer: ProducerID,
				ClaimValid: true, ClaimNote: "fixture asserts valid",
			}, nil
		case FixtureClaimPass:
			raw, err := json.Marshal(map[string]any{
				"schema": schemaID,
				"title":  "broken",
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

func draftValid(obj kernel.Objective) productionPackage {
	title := "Demo package"
	if obj.Text != "" {
		title = "Package for: " + obj.Text
	}
	return productionPackage{
		Schema:   schemaID,
		Title:    title,
		Audience: "operators",
		Message:  "prove governed production",
		Scenes: []scene{
			{ID: "2", Outline: "close"},
			{ID: "1", Outline: "open"},
		},
		Narration: "A short narration fragment.",
		Assets: []asset{
			{Name: "b-roll"},
			{Name: "a-roll"},
		},
	}
}

// Transform canonicalizes a candidate. Same input bytes → same output bytes.
func Transform(in []byte) ([]byte, error) {
	var pkg productionPackage
	if err := json.Unmarshal(in, &pkg); err != nil {
		return nil, fmt.Errorf("candidate schema decode: %w", err)
	}
	if pkg.Schema == "" {
		pkg.Schema = schemaID
	}
	sort.Slice(pkg.Scenes, func(i, j int) bool { return pkg.Scenes[i].ID < pkg.Scenes[j].ID })
	sort.Slice(pkg.Assets, func(i, j int) bool { return pkg.Assets[i].Name < pkg.Assets[j].Name })
	for i := range pkg.Assets {
		pkg.Assets[i].Hash = kernel.HashBytes([]byte(pkg.Title + "|" + pkg.Assets[i].Name))
	}
	return json.Marshal(pkg)
}

// Verify checks the transformed artifact against pinned criteria JSON.
func Verify(artifact, criteriaJSON []byte, criteriaHash string) (kernel.Verdict, error) {
	v := kernel.Verdict{
		Evaluator:    VerifierID,
		CriteriaHash: criteriaHash,
		ArtifactHash: kernel.HashBytes(artifact),
	}
	var crit criteria
	if err := json.Unmarshal(criteriaJSON, &crit); err != nil {
		v.Note = "criteria decode failed"
		return v, nil
	}
	var pkg productionPackage
	if err := json.Unmarshal(artifact, &pkg); err != nil {
		v.Note = "artifact is not a production package"
		v.Evidence, _ = json.Marshal(map[string]string{"error": err.Error()})
		return v, nil
	}

	var failures []string
	if pkg.Schema != schemaID {
		failures = append(failures, "schema mismatch")
	}
	fieldEmpty := map[string]bool{
		"title":     strings.TrimSpace(pkg.Title) == "",
		"audience":  strings.TrimSpace(pkg.Audience) == "",
		"message":   strings.TrimSpace(pkg.Message) == "",
		"scenes":    len(pkg.Scenes) == 0,
		"narration": strings.TrimSpace(pkg.Narration) == "",
		"assets":    len(pkg.Assets) == 0,
	}
	for _, f := range crit.RequiredFields {
		if fieldEmpty[f] {
			failures = append(failures, "missing "+f)
		}
	}
	if crit.MinScenes > 0 && len(pkg.Scenes) < crit.MinScenes {
		failures = append(failures, "too few scenes")
	}
	if crit.RequireSortedScenes {
		for i := 1; i < len(pkg.Scenes); i++ {
			if pkg.Scenes[i-1].ID > pkg.Scenes[i].ID {
				failures = append(failures, "scenes not sorted")
				break
			}
		}
	}
	if crit.RequireSortedAssets {
		for i := 1; i < len(pkg.Assets); i++ {
			if pkg.Assets[i-1].Name > pkg.Assets[i].Name {
				failures = append(failures, "assets not sorted")
				break
			}
		}
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

	v.Evidence, _ = json.Marshal(map[string]any{
		"failures": failures, "schema": pkg.Schema, "criteria_hash": criteriaHash,
	})
	if len(failures) > 0 {
		v.Note = strings.Join(failures, "; ")
		return v, nil
	}
	v.Pass = true
	v.Note = "criteria satisfied"
	return v, nil
}
