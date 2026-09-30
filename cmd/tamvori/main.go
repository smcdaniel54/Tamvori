package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "usage: tamvori run [--fixture valid|invalid|claim_pass] [--workdir DIR]\n")
		os.Exit(2)
	}
	fixture := media.FixtureValid
	workDir := filepath.Join(".", ".tamvori-run")
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--fixture":
			i++
			if i >= len(os.Args) {
				fatal("missing --fixture value")
			}
			fixture = media.FixtureMode(os.Args[i])
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

	criteriaPath := filepath.Join("experiments", "exp-001b-kernel-slice", "acceptance", "criteria.json")
	criteria, err := os.ReadFile(criteriaPath)
	if err != nil {
		fatal(err.Error())
	}

	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir:      workDir,
		Objective:    kernel.Objective{ID: "obj-001", Text: "produce a demo media package"},
		CriteriaJSON: criteria,
		Bounds:       kernel.Bounds{MaxAttempts: 2, MaxBytes: 1 << 20},
		Propose:      media.Propose(fixture),
		Transform:    media.Transform,
		Verify:       media.Verify,
		VerifierID:   media.VerifierID,
	})
	if err != nil {
		fatal(err.Error())
	}
	out, _ := json.MarshalIndent(map[string]any{
		"status":          res.Status,
		"criteria_hash":   res.CriteriaHash,
		"plan_hash":       res.PlanHash,
		"candidate_hash":  res.CandidateHash,
		"artifact_hash":   res.ArtifactHash,
		"reject_reason":   res.RejectReason,
		"event_count":     len(res.Events),
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
