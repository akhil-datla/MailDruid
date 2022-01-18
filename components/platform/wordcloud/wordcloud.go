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

var DefaultColors = []color.RGBA{
	{0x1b, 0x1b, 0x1b, 0xff},
	{0x48, 0x48, 0x4B, 0xff},
	{0x59, 0x3a, 0xee, 0xff},
	{0x65, 0xCD, 0xFA, 0xff},
	{0x70, 0xD6, 0xBF, 0xff},
}

func Summarize(text string, maxLength int) string {
	bag := tldr.New()
	result, _ := bag.Summarize(text, maxLength)

	var summary string
	for _, sentence := range result {
		summary += sentence + " "
	}
	return summary
}

func ExtractKeyWords(text string) map[string]int {

	candidates := rake.RunRake(text)

	keywords := make(map[string]int)

	for _, candidate := range candidates {
		keywords[candidate.Key] = int(candidate.Value)
	}
	return keywords
}

func GenerateWordCloud(wordCounts map[string]int) (string, error) {

	colors := make([]color.Color, 0)
	for _, c := range DefaultColors {
		colors = append(colors, c)
	}

	fontFilePath := filepath.Join(filepath.Dir(""), "fonts", "roboto", "Roboto-Regular.ttf")

	w := wordclouds.NewWordcloud(
		wordCounts,
		wordclouds.FontFile(fontFilePath),
		wordclouds.Height(2048),
		wordclouds.Width(2048),
		wordclouds.Colors(colors),
		wordclouds.RandomPlacement(false),
	)

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	img := w.Draw()

	fileName := fmt.Sprintf("%s.png", id.String())

	outputFile, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	png.Encode(outputFile, img)

	defer outputFile.Close()

	return fileName, nil

}
