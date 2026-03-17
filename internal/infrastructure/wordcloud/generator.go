package wordcloud

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/JesusIslam/tldr"
	rake "github.com/afjoseph/RAKE.go"
	"github.com/gofrs/uuid"
	"github.com/psykhi/wordclouds"
)

var defaultColors = []color.RGBA{
	{0x1B, 0x1B, 0x1B, 0xFF},
	{0x48, 0x48, 0x4B, 0xFF},
	{0x59, 0x3A, 0xEE, 0xFF},
	{0x65, 0xCD, 0xFA, 0xFF},
	{0x70, 0xD6, 0xBF, 0xFF},
}

// Generator creates text summaries and word cloud visualizations.
type Generator struct {
	fontPath string
}

// New creates a word cloud generator.
// fontPath should point to a TTF font file.
func New(fontPath string) *Generator {
	return &Generator{fontPath: fontPath}
}

// Summarize extracts the top sentences from text.
func (g *Generator) Summarize(text string, maxSentences int) string {
	if maxSentences <= 0 {
		maxSentences = 5
	}

	bag := tldr.New()
	result, _ := bag.Summarize(text, maxSentences)

	var summary string
	for _, sentence := range result {
		summary += sentence + " "
	}
	return summary
}

// ExtractKeywords uses RAKE to extract keywords and their scores.
func (g *Generator) ExtractKeywords(text string) map[string]int {
	candidates := rake.RunRake(text)
	keywords := make(map[string]int, len(candidates))
	for _, c := range candidates {
		keywords[c.Key] = int(c.Value)
	}
	return keywords
}

// GenerateWordCloud creates a PNG word cloud and returns the file path.
func (g *Generator) GenerateWordCloud(wordCounts map[string]int) (string, error) {
	if len(wordCounts) == 0 {
		return "", nil
	}

	colors := make([]color.Color, len(defaultColors))
	for i, c := range defaultColors {
		colors[i] = c
	}

	w := wordclouds.NewWordcloud(
		wordCounts,
		wordclouds.FontFile(g.fontPath),
		wordclouds.Height(2048),
		wordclouds.Width(2048),
		wordclouds.Colors(colors),
		wordclouds.RandomPlacement(false),
	)

	img := w.Draw()

	id, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("generating filename: %w", err)
	}

	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("maildruid-%s.png", id.String()))
	f, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		os.Remove(fileName)
		return "", fmt.Errorf("encoding PNG: %w", err)
	}

	return fileName, nil
}
