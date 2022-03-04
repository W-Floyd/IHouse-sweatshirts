package main

import (
	"image"
	"log"
	"strings"

	"github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
)

type Text struct {
	Text     string
	Scale    float64
	Typeface string
	CaseUp   bool
	CaseDown bool
}

type RenderOptions struct {
	GlobalScale      float64
	SourceWidth      int
	DefaultFontScale float64
	DefaultTypeface  string
	BaseImage        image.RGBA
}

var GlobalOptions = RenderOptions{
	GlobalScale:      0.125,
	DefaultFontScale: 1 / 6.25,
	DefaultTypeface:  "College.ttf",
}

func (text *Text) TransformText() string {
	if text.CaseUp {
		return strings.ToUpper(text.Text)
	}
	if text.CaseDown {
		return strings.ToLower(text.Text)
	}
	return text.Text
}

func (renderOptions RenderOptions) GetBaseFontSize() float64 {
	return float64(GlobalOptions.SourceWidth) * renderOptions.DefaultFontScale
}

func (output Output) splitText() (retval []Text) {

	for _, substring := range strings.Split(output.Text, "\\n") {
		trimmed := strings.TrimSpace(substring)

		if trimmed == "" {
			continue
		}

		scale := 1.0

		oldTrimmed := trimmed

		typeface := GlobalOptions.DefaultTypeface

		caseUp := true
		caseDown := false

		for {
			if strings.HasSuffix(trimmed, ":small") {
				trimmed = strings.TrimSuffix(trimmed, ":small")
				scale = 0.5
			}

			if strings.HasSuffix(trimmed, ":large") {
				trimmed = strings.TrimSuffix(trimmed, ":large")
				scale = 1.5
			}

			if strings.HasSuffix(trimmed, ":lower") {
				trimmed = strings.TrimSuffix(trimmed, ":lower")
				caseUp = false
				caseDown = true
			}

			if strings.HasSuffix(trimmed, ":upper") {
				trimmed = strings.TrimSuffix(trimmed, ":upper")
				caseUp = true
				caseDown = false
			}

			if strings.HasSuffix(trimmed, ":nocase") {
				trimmed = strings.TrimSuffix(trimmed, ":nocase")
				caseUp = false
				caseDown = false
			}

			if strings.HasSuffix(trimmed, ":serif") {
				trimmed = strings.TrimSuffix(trimmed, ":serif")
				typeface = "FreeSerif.ttf"
			}

			if strings.HasSuffix(trimmed, ":sans") {
				trimmed = strings.TrimSuffix(trimmed, ":sans")
				typeface = "LiberationSans.ttf"
			}

			if trimmed == oldTrimmed {
				break
			}

			oldTrimmed = trimmed
		}

		retval = append(retval, Text{
			Text:     trimmed,
			Scale:    scale,
			Typeface: typeface,
			CaseUp:   caseUp,
			CaseDown: caseDown,
		})

	}

	return
}

func (text Text) renderText() image.Image {
	c := gg.NewContext(int(GlobalOptions.SourceWidth), int(GlobalOptions.DefaultFontScale*float64(GlobalOptions.SourceWidth)))
	s := 1.0
	fontPath, err := findfont.Find(text.Typeface)
	if err != nil {
		log.Fatalln(err)
	}
	c.LoadFontFace(fontPath, s*text.Scale*GlobalOptions.GetBaseFontSize())
	if w, _ := c.MeasureString(text.TransformText()); w > float64(GlobalOptions.SourceWidth) {
		s = float64(GlobalOptions.SourceWidth) / w
	}

	_, h := c.MeasureMultilineString(text.TransformText(), 0)

	c = gg.NewContext(int(GlobalOptions.SourceWidth), int(h*1.05))
	c.LoadFontFace(fontPath, s*text.Scale*GlobalOptions.GetBaseFontSize())

	c.SetRGBA(0, 0, 0, 1)
	c.DrawStringAnchored(text.TransformText(), float64(c.Width()/2), float64(c.Height()/2), 0.5, 0.5)
	return c.Image()
}

func stackText(texts []Text) image.Image {
	images := []image.Image{}
	height := 0
	y := []int{}
	for _, text := range texts {
		im := text.renderText()
		height = height + im.Bounds().Bounds().Dy()
		images = append(images, im)
		y = append(y, height)
	}

	c := gg.NewContext(GlobalOptions.SourceWidth, height)
	yval := 0
	for i := range y {
		c.SetRGBA(0, 0, 0, 1)
		c.DrawImage(images[i], 0, yval)
		yval = y[i]
	}
	return c.Image()
}
