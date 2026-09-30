package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "usage: tamvori run [--fixture valid|invalid|claim_pass | --live] [--model NAME] [--workdir DIR]\n")
		os.Exit(2)
	}
	fixture := media.FixtureValid
	live := false
	model := media.DefaultLiveModel
	workDir := filepath.Join(".", ".tamvori-run")
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--fixture":
			i++
			if i >= len(os.Args) {
				fatal("missing --fixture value")
			}
			fixture = os.Args[i]
		case "--live":
			live = true
		case "--model":
			i++
			if i >= len(os.Args) {
				fatal("missing --model value")
			}
			model = os.Args[i]
		case "--workdir":
			i++
			if i >= len(os.Args) {
				fatal("missing --workdir value")
			}
			workDir = os.Args[i]
		default:
			fatal("unknown flag " + os.Args[i])
		}
	}

	criteriaPath := filepath.Join("acceptance", "media-package.json")
	criteria, err := os.ReadFile(criteriaPath)
	if err != nil {
		fatal(err.Error())
	}

	propose := media.Propose(fixture)
	bounds := kernel.Bounds{MaxAttempts: 2, MaxBytes: 1 << 20}
	obj := kernel.Objective{ID: "obj-001", Text: "produce a demo media package"}
	if live {
		propose = media.ProposeLive(media.LiveConfig{
			Model:   model,
			Timeout: 45 * time.Second,
		})
		bounds.MaxAttempts = 1 // EXP-001C: no live retries
		obj = kernel.Objective{
			ID:   "obj-001c-live",
			Text: "Create a short structured media production package about proving governed AI generation for operators.",
		}
	}

	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      workDir,
		Objective:    obj,
		CriteriaJSON: criteria,
		Bounds:       bounds,
		Propose:      propose,
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		fatal(err.Error())
	}
	out, _ := json.MarshalIndent(map[string]any{
		"status":         res.Status,
		"criteria_hash":  res.CriteriaHash,
		"plan_hash":      res.PlanHash,
		"candidate_hash": res.CandidateHash,
		"artifact_hash":  res.ArtifactHash,
		"reject_reason":  res.RejectReason,
		"event_count":    len(res.Events),
		"live":           live,
	}, "", "  ")
	fmt.Println(string(out))
	if res.Status != "accepted" {
		os.Exit(1)
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
