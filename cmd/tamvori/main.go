package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
	"github.com/smcdaniel54/Tamvori/internal/render"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "run":
		cmdRun(os.Args[2:])
	case "render":
		cmdRender(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage:\n  tamvori run [--fixture valid|invalid|claim_pass | --live] [--model NAME] [--workdir DIR]\n  tamvori render --input FILE [--workdir DIR] [--ffmpeg PATH] [--ffprobe PATH]\n")
	os.Exit(2)
}

func cmdRun(args []string) {
	fixture := media.FixtureValid
	live := false
	model := media.DefaultLiveModel
	workDir := filepath.Join(".", ".tamvori-run")
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--fixture":
			i++
			if i >= len(args) {
				fatal("missing --fixture value")
			}
			fixture = args[i]
		case "--live":
			live = true
		case "--model":
			i++
			if i >= len(args) {
				fatal("missing --model value")
			}
			model = args[i]
		case "--workdir":
			i++
			if i >= len(args) {
				fatal("missing --workdir value")
			}
			workDir = args[i]
		default:
			fatal("unknown flag " + args[i])
		}
	}

	criteria, err := os.ReadFile(filepath.Join("acceptance", "media-package.json"))
	if err != nil {
		fatal(err.Error())
	}

	propose := media.Propose(fixture)
	bounds := kernel.Bounds{MaxAttempts: 2, MaxBytes: 1 << 20}
	obj := kernel.Objective{ID: "obj-001", Text: "produce a demo media package"}
	if live {
		propose = media.ProposeLive(media.LiveConfig{Model: model, Timeout: 45 * time.Second})
		bounds.MaxAttempts = 1
		obj = kernel.Objective{
			ID:   "obj-001c-live",
			Text: "Create a short structured media production package about proving governed AI generation for operators.",
		}
	}

	res, err := kernel.Execute(kernel.RunRequest{
		WorkDir: workDir, Objective: obj, CriteriaJSON: criteria, Bounds: bounds,
		Propose: propose, Transform: media.Transform, Verify: media.Verify, VerifierID: media.VerifierID,
	})
	if err != nil {
		fatal(err.Error())
	}
	printJSON(map[string]any{
		"status": res.Status, "criteria_hash": res.CriteriaHash, "plan_hash": res.PlanHash,
		"candidate_hash": res.CandidateHash, "artifact_hash": res.ArtifactHash,
		"reject_reason": res.RejectReason, "event_count": len(res.Events), "live": live,
	})
	if res.Status != "accepted" {
		os.Exit(1)
	}
}

func cmdRender(args []string) {
	input := ""
	workDir := filepath.Join(".", ".tamvori-render")
	ffmpegPath := "ffmpeg"
	ffprobePath := "ffprobe"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--input":
			i++
			if i >= len(args) {
				fatal("missing --input value")
			}
			input = args[i]
		case "--workdir":
			i++
			if i >= len(args) {
				fatal("missing --workdir value")
			}
			workDir = args[i]
		case "--ffmpeg":
			i++
			if i >= len(args) {
				fatal("missing --ffmpeg value")
			}
			ffmpegPath = args[i]
		case "--ffprobe":
			i++
			if i >= len(args) {
				fatal("missing --ffprobe value")
			}
			ffprobePath = args[i]
		default:
			fatal("unknown flag " + args[i])
		}
	}
	if input == "" {
		fatal("--input required")
	}
	pkg, err := os.ReadFile(input)
	if err != nil {
		fatal(err.Error())
	}
	criteria, err := os.ReadFile(filepath.Join("acceptance", "media-package.json"))
	if err != nil {
		fatal(err.Error())
	}
	res, err := render.AuthorizeAndRender(context.Background(), pkg, criteria, render.Config{
		WorkDir: workDir, FFmpegPath: ffmpegPath, FFprobePath: ffprobePath,
		Timeout: render.DefaultTimeout,
	})
	if err != nil {
		fatal(err.Error())
	}
	printJSON(map[string]any{
		"authorized":     res.Authorized,
		"source_hash":    res.SourceHash,
		"plan_hash":      res.PlanHash,
		"output_path":    res.OutputPath,
		"output_hash":    res.OutputHash,
		"output_bytes":   res.OutputBytes,
		"ffmpeg_version": res.FFmpegVersion,
		"verify_notes":   res.VerifyNotes,
		"reject_reason":  res.RejectReason,
	})
	if !res.Authorized || res.OutputHash == "" {
		os.Exit(1)
	}
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
