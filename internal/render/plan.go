package render

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Deterministic defaults for EXP-001D (boring proof video).
const (
	Width         = 1280
	Height        = 720
	FPS           = 30
	SceneDuration = 2 // seconds per card
	Format        = "mp4"
	RendererID    = "render/ffmpeg-card"
)

// Plan is the minimal deterministic render input derived from an accepted package.
type Plan struct {
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	FPS           int      `json:"fps"`
	SceneDuration int      `json:"scene_duration_sec"`
	Format        string   `json:"format"`
	Cards         []string `json:"cards"` // display text, one card per scene (+ title card)
	SourceHash    string   `json:"source_package_hash"`
}

type packageView struct {
	Title     string `json:"title"`
	Audience  string `json:"audience"`
	Message   string `json:"message"`
	Narration string `json:"narration"`
	Scenes    []struct {
		ID      string `json:"id"`
		Outline string `json:"outline"`
	} `json:"scenes"`
}

// BuildPlan creates a deterministic render plan from accepted package bytes.
func BuildPlan(packageJSON []byte) (Plan, error) {
	var pkg packageView
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		return Plan{}, fmt.Errorf("render plan: %w", err)
	}
	sum := sha256.Sum256(packageJSON)
	sourceHash := hex.EncodeToString(sum[:])

	cards := []string{
		safeCard("TITLE", pkg.Title),
		safeCard("AUDIENCE", pkg.Audience),
		safeCard("MESSAGE", pkg.Message),
	}
	for _, s := range pkg.Scenes {
		cards = append(cards, safeCard(s.ID, s.Outline))
	}
	if strings.TrimSpace(pkg.Narration) != "" {
		cards = append(cards, safeCard("NARRATION", pkg.Narration))
	}
	if len(cards) == 0 {
		return Plan{}, fmt.Errorf("render plan: no cards")
	}
	return Plan{
		Width: Width, Height: Height, FPS: FPS,
		SceneDuration: SceneDuration, Format: Format,
		Cards: cards, SourceHash: sourceHash,
	}, nil
}

func (p Plan) Hash() (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// safeCard removes control characters and limits length for drawtext files.
// It never becomes a shell argument — text is written to a file.
func safeCard(label, body string) string {
	label = scrub(label)
	body = scrub(body)
	if label == "" {
		label = "CARD"
	}
	line := label + ": " + body
	if len(line) > 180 {
		line = line[:177] + "..."
	}
	return line
}

func scrub(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			continue
		}
		if r == '\n' || r == '\r' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// DiscoverFont returns an existing system font path suitable for drawtext, or "".
func DiscoverFont() string {
	candidates := []string{
		filepath.Join(os.Getenv("WINDIR"), "Fonts", "arial.ttf"),
		filepath.Join(os.Getenv("WINDIR"), "Fonts", "segoeui.ttf"),
		filepath.Join(os.Getenv("WINDIR"), "Fonts", "calibri.ttf"),
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/Supplemental/Arial.ttf",
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}
