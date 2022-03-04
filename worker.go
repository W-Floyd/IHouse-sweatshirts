package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"sync"

	"github.com/fogleman/gg"
)

var wg sync.WaitGroup

func worker(jobs <-chan Output) {

	for outputEntry := range jobs {

		enc := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}

		outputPath := "output/" + outputEntry.filename() + ".png"

		textImage := stackText(outputEntry.splitText())

		height := textImage.Bounds().Max.Y

		textY := 0

		if outputEntry.IncludeImage {
			height = height + GlobalOptions.BaseImage.Bounds().Dy()
		}

		dc := gg.NewContext(GlobalOptions.SourceWidth, height)

		if outputEntry.IncludeImage {
			dc.DrawImage(&GlobalOptions.BaseImage, 0, 0)
			textY = GlobalOptions.BaseImage.Bounds().Dy()
		}

		dc.DrawImage(textImage, 0, textY)

		file, err := os.Create(outputPath)
		defer file.Close()
		if err != nil {
			log.Fatalln(err)
		}

		enc.Encode(file, dc.Image())
		fmt.Println("Done: " + file.Name())

		wg.Done()
	}
}
