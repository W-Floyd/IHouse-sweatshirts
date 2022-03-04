package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"runtime"

	"github.com/nfnt/resize"
	"github.com/xuri/excelize/v2"
)

type Entry struct {
	Name       string
	ShirtSize  string
	ShirtColor string
	LogoSize   string
	LogoSide   string
	TextFront  string
	TextRear   string
}

type Entries []Entry

func (output Output) filename() string {
	return output.Entry.Name +
		" - " +
		output.Entry.ShirtSize +
		" - " +
		output.Entry.ShirtColor +
		" - " +
		func(output Output) string {
			if output.IncludeImage {
				return output.Entry.LogoSize + " - "
			}
			return ""
		}(output) +
		output.Side

}

type Output struct {
	Side         string
	Entry        Entry
	IncludeImage bool
	Text         string
}

type Outputs []Output

func (entries Entries) process() (output Outputs) {
	for _, entry := range entries {
		if (entry.TextFront != "") || (entry.LogoSide == "Front") {
			output = append(output, Output{
				Side:         "Front",
				Entry:        entry,
				Text:         entry.TextFront,
				IncludeImage: entry.LogoSide == "Front",
			})
		}
		if (entry.TextRear != "") || (entry.LogoSide == "Rear") {
			output = append(output, Output{
				Side:         "Rear",
				Entry:        entry,
				Text:         entry.TextRear,
				IncludeImage: entry.LogoSide == "Rear",
			})
		}
	}

	return
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func main() {

	f, err := excelize.OpenFile("Sweatshirts.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return
	}

	entries := Entries{}

	for i, row := range rows {
		if (i > 0) && (len(row) > 0) {

			r := row

			for i := 0; i < 7-len(row); i++ {
				r = append(r, "")
			}

			e := Entry{
				Name:       r[0],
				ShirtSize:  r[1],
				ShirtColor: r[2],
				LogoSize:   r[3],
				LogoSide:   r[4],
				TextFront:  r[5],
				TextRear:   r[6],
			}

			if e.TextFront == "-" {
				e.TextFront = ""
			}

			if e.TextRear == "-" {
				e.TextRear = ""
			}

			entries = append(entries, e)

		}
	}

	outputs := entries.process()

	inputImagePath := "Logo_BW_flat.png"

	fmt.Println("Loading source image")

	file, err := os.Open(inputImagePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	src, _, err := image.Decode(file)

	src = resize.Resize(
		uint(float64(src.Bounds().Dx())*GlobalOptions.GlobalScale),
		uint(float64(src.Bounds().Dy())*GlobalOptions.GlobalScale),
		src,
		resize.NearestNeighbor,
	)

	bounds := src.Bounds()
	GlobalOptions.BaseImage = *image.NewRGBA(bounds)
	draw.Draw(&GlobalOptions.BaseImage, bounds, src, bounds.Min, draw.Src)

	GlobalOptions.SourceWidth = bounds.Dx()

	if err != nil {
		log.Fatalln(err)
	}

	jobs := make(chan Output, len(outputs))

	for w := 0; w < runtime.NumCPU()*2; w++ {
		go worker(jobs)
	}

	fmt.Println("Queuing jobs...")
	for _, output := range outputs {
		wg.Add(1)
		jobs <- output
	}

	close(jobs)

	wg.Wait()

}
