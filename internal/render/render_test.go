package render_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/smcdaniel54/Tamvori/internal/render"
)

func criteria(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "acceptance", "media-package.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func validPkg() []byte {
	return []byte(`{
	  "schema":"tamvori.media.production_package.v1",
	  "title":"Render Control",
	  "audience":"operators",
	  "message":"proof",
	  "scenes":[{"id":"1","outline":"open"},{"id":"2","outline":"close"}],
	  "narration":"A short line.",
	  "assets":[{"name":"a-roll","hash":""}]
	}`)
}

func TestRejectedPackageBlocksRender(t *testing.T) {
	bad := []byte(`{"schema":"tamvori.media.production_package.v1","title":"broken"}`)
	res, err := render.AuthorizeAndRender(context.Background(), bad, criteria(t), render.Config{
		WorkDir:    t.TempDir(),
		FFmpegPath: filepath.Join(t.TempDir(), "must-not-run"),
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Authorized {
		t.Fatal("rejected package must not authorize render")
	}
	if !strings.Contains(res.RejectReason, "render not authorized") {
		t.Fatalf("reason=%q", res.RejectReason)
	}
}

func TestBuildPlanDeterministic(t *testing.T) {
	a, err := render.BuildPlan(validPkg())
	if err != nil {
		t.Fatal(err)
	}
	b, err := render.BuildPlan(validPkg())
	if err != nil {
		t.Fatal(err)
	}
	ha, _ := a.Hash()
	hb, _ := b.Hash()
	if ha != hb {
		t.Fatal("plan hash mismatch")
	}
	if a.Width != 1280 || a.Height != 720 || a.FPS != 30 {
		t.Fatalf("unexpected defaults %+v", a)
	}
}

func TestMissingFFmpegGoverned(t *testing.T) {
	plan, err := render.BuildPlan(validPkg())
	if err != nil {
		t.Fatal(err)
	}
	_, err = render.Execute(context.Background(), plan, render.Config{
		WorkDir:    t.TempDir(),
		FFmpegPath: filepath.Join(t.TempDir(), "no-such-ffmpeg"),
		FontPath:   mustFont(t),
		Timeout:    2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected missing ffmpeg failure")
	}
}

func TestFakeFFmpegTimeout(t *testing.T) {
	fake := buildSleepFake(t)
	plan, err := render.BuildPlan(validPkg())
	if err != nil {
		t.Fatal(err)
	}
	_, err = render.Execute(context.Background(), plan, render.Config{
		WorkDir:    t.TempDir(),
		FFmpegPath: fake,
		FontPath:   mustFont(t),
		Timeout:    200 * time.Millisecond,
	})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestFakeFFmpegNonZeroExit(t *testing.T) {
	fake := buildFailFake(t)
	plan, err := render.BuildPlan(validPkg())
	if err != nil {
		t.Fatal(err)
	}
	_, err = render.Execute(context.Background(), plan, render.Config{
		WorkDir:    t.TempDir(),
		FFmpegPath: fake,
		FontPath:   mustFont(t),
		Timeout:    2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected non-zero exit failure")
	}
}

func mustFont(t *testing.T) string {
	t.Helper()
	f := render.DiscoverFont()
	if f == "" {
		t.Skip("no system font")
	}
	return f
}

func buildSleepFake(t *testing.T) string {
	t.Helper()
	src := `package main
import ("os"; "time")
func main() { time.Sleep(2 * time.Second); os.Exit(0) }`
	return compileFake(t, "sleepfake", src)
}

func buildFailFake(t *testing.T) string {
	t.Helper()
	src := `package main
import "os"
func main() { os.Exit(2) }`
	return compileFake(t, "failfake", src)
}

func compileFake(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", out, srcPath)
	cmd.Env = os.Environ()
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake: %v\n%s", err, b)
	}
	return out
}
