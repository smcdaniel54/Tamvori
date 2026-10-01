package render

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/kernel"
	"github.com/smcdaniel54/Tamvori/internal/media"
)

// DefaultTimeout bounds one FFmpeg render.
const DefaultTimeout = 120 * time.Second

// Config controls one concrete FFmpeg render. No renderer registry.
type Config struct {
	FFmpegPath  string
	FFprobePath string
	FontPath    string
	WorkDir     string
	Timeout     time.Duration
}

// Result is the outcome of an authorized render attempt.
type Result struct {
	Authorized    bool
	SourceHash    string
	PlanHash      string
	OutputPath    string
	OutputHash    string
	OutputBytes   int64
	FFmpegVersion string
	VerifyNotes   []string
	RejectReason  string
}

// AuthorizeAndRender verifies the package against pinned criteria, then renders.
// Rejected packages never invoke FFmpeg.
func AuthorizeAndRender(ctx context.Context, packageJSON, criteriaJSON []byte, cfg Config) (Result, error) {
	transformed, err := media.Transform(packageJSON)
	if err != nil {
		return Result{RejectReason: "transform: " + err.Error()}, nil
	}
	criteriaHash := kernel.HashBytes(criteriaJSON)
	v, err := media.Verify(transformed, criteriaJSON, criteriaHash)
	if err != nil {
		return Result{}, err
	}
	if !v.Pass {
		return Result{
			Authorized:   false,
			SourceHash:   kernel.HashBytes(transformed),
			RejectReason: "render not authorized: " + v.Note,
		}, nil
	}

	plan, err := BuildPlan(transformed)
	if err != nil {
		return Result{Authorized: true, SourceHash: plan.SourceHash, RejectReason: err.Error()}, nil
	}
	planHash, err := plan.Hash()
	if err != nil {
		return Result{}, err
	}

	res, err := Execute(ctx, plan, cfg)
	if err != nil {
		return Result{
			Authorized: true, SourceHash: plan.SourceHash, PlanHash: planHash,
			RejectReason: err.Error(),
		}, nil
	}
	res.Authorized = true
	res.SourceHash = plan.SourceHash
	res.PlanHash = planHash
	return res, nil
}

// Execute runs FFmpeg for an already-authorized plan. Tests may inject a fake binary.
func Execute(ctx context.Context, plan Plan, cfg Config) (Result, error) {
	if cfg.WorkDir == "" {
		return Result{}, fmt.Errorf("render: workdir required")
	}
	if err := os.MkdirAll(cfg.WorkDir, 0o755); err != nil {
		return Result{}, err
	}
	ffmpeg := cfg.FFmpegPath
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	font := cfg.FontPath
	if font == "" {
		font = DiscoverFont()
	}
	if font == "" {
		return Result{}, fmt.Errorf("render: no system font found")
	}
	// Copy font into workdir so the filter path has no drive-letter colon (Windows).
	// The font is not committed to the repository.
	localFont := filepath.Join(cfg.WorkDir, "font.ttf")
	if err := copyFile(font, localFont); err != nil {
		return Result{}, fmt.Errorf("render: prepare font: %w", err)
	}
	font = localFont
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	version, _ := ffmpegVersion(ctx, ffmpeg, cfg.WorkDir)

	scenesDir := filepath.Join(cfg.WorkDir, "scenes")
	if err := os.MkdirAll(scenesDir, 0o755); err != nil {
		return Result{}, err
	}

	var sceneFiles []string
	for i, card := range plan.Cards {
		textRel := filepath.Join("scenes", fmt.Sprintf("card_%02d.txt", i))
		textPath := filepath.Join(cfg.WorkDir, textRel)
		if err := os.WriteFile(textPath, []byte(card), 0o644); err != nil {
			return Result{}, err
		}
		outRel := filepath.Join("scenes", fmt.Sprintf("scene_%02d.mp4", i))
		// Relative paths inside workdir avoid Windows drive-letter filter escaping issues.
		vf := fmt.Sprintf(
			"drawtext=fontfile=%s:textfile=%s:fontsize=36:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2",
			ffmpegFilterPath("font.ttf"),
			ffmpegFilterPath(filepath.ToSlash(textRel)),
		)
		args := []string{
			"-y",
			"-hide_banner", "-loglevel", "error",
			"-f", "lavfi",
			"-i", fmt.Sprintf("color=c=black:s=%dx%d:d=%d:r=%d", plan.Width, plan.Height, plan.SceneDuration, plan.FPS),
			"-vf", vf,
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-an",
			"-fflags", "+bitexact",
			"-flags:v", "+bitexact",
			"-map_metadata", "-1",
			"-movflags", "+faststart",
			outRel,
		}
		if err := run(ctx, ffmpeg, args, cfg.WorkDir); err != nil {
			return Result{FFmpegVersion: version}, fmt.Errorf("ffmpeg scene %d: %w", i, err)
		}
		sceneFiles = append(sceneFiles, outRel)
	}

	listPath := filepath.Join(cfg.WorkDir, "concat.txt")
	var list bytes.Buffer
	for _, f := range sceneFiles {
		p := strings.ReplaceAll(filepath.ToSlash(f), "'", `'\''`)
		fmt.Fprintf(&list, "file '%s'\n", p)
	}
	if err := os.WriteFile(listPath, list.Bytes(), 0o644); err != nil {
		return Result{}, err
	}

	finalName := "render_" + plan.SourceHash[:16] + ".mp4"
	finalPath := filepath.Join(cfg.WorkDir, finalName)
	if !underDir(cfg.WorkDir, finalPath) {
		return Result{}, fmt.Errorf("render: output path escapes workdir")
	}

	concatArgs := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-f", "concat", "-safe", "0",
		"-i", "concat.txt",
		"-c", "copy",
		"-fflags", "+bitexact",
		"-map_metadata", "-1",
		"-movflags", "+faststart",
		finalName,
	}
	if err := run(ctx, ffmpeg, concatArgs, cfg.WorkDir); err != nil {
		return Result{FFmpegVersion: version}, fmt.Errorf("ffmpeg concat: %w", err)
	}

	st, err := os.Stat(finalPath)
	if err != nil {
		return Result{FFmpegVersion: version}, fmt.Errorf("render: output missing: %w", err)
	}
	if st.Size() == 0 {
		return Result{FFmpegVersion: version}, fmt.Errorf("render: output empty")
	}
	raw, err := os.ReadFile(finalPath)
	if err != nil {
		return Result{}, err
	}
	sum := sha256.Sum256(raw)
	outHash := hex.EncodeToString(sum[:])

	notes := []string{"exit=0", "nonempty=true", "ext=mp4"}
	ffprobe := cfg.FFprobePath
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	if props, err := probe(ctx, ffprobe, finalPath); err == nil {
		notes = append(notes, props...)
	} else {
		notes = append(notes, "ffprobe="+err.Error())
	}

	return Result{
		OutputPath:    finalPath,
		OutputHash:    outHash,
		OutputBytes:   st.Size(),
		FFmpegVersion: version,
		VerifyNotes:   notes,
	}, nil
}

func run(ctx context.Context, bin string, args []string, dir string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout: %s", msg)
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func ffmpegVersion(ctx context.Context, bin, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, "-version")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	line := strings.SplitN(string(out), "\n", 2)[0]
	return strings.TrimSpace(line), nil
}

func probe(ctx context.Context, bin, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, bin,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,avg_frame_rate,codec_type:format=duration,format_name",
		"-of", "json",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Streams []struct {
			CodecType    string `json:"codec_type"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			AvgFrameRate string `json:"avg_frame_rate"`
		} `json:"streams"`
		Format struct {
			Duration   string `json:"duration"`
			FormatName string `json:"format_name"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, err
	}
	notes := []string{}
	if len(parsed.Streams) > 0 {
		s := parsed.Streams[0]
		notes = append(notes,
			"codec="+s.CodecType,
			"width="+strconv.Itoa(s.Width),
			"height="+strconv.Itoa(s.Height),
			"fps="+s.AvgFrameRate,
		)
		if s.Width != Width || s.Height != Height {
			return notes, fmt.Errorf("unexpected dimensions %dx%d", s.Width, s.Height)
		}
	}
	notes = append(notes, "format="+parsed.Format.FormatName, "duration="+parsed.Format.Duration)
	return notes, nil
}

// ffmpegFilterPath converts a relative workdir path for use inside an FFmpeg filtergraph.
func ffmpegFilterPath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.ReplaceAll(p, "'", `\'`)
	p = strings.ReplaceAll(p, ":", `\:`)
	p = strings.ReplaceAll(p, "[", `\[`)
	p = strings.ReplaceAll(p, "]", `\]`)
	return p
}

func underDir(dir, path string) bool {
	absDir, err1 := filepath.Abs(dir)
	absPath, err2 := filepath.Abs(path)
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

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}
