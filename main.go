package main

import (
	"time"
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/alexa-infra/kalendar/calendar"
)

var (
	imgfile = flag.String("imgfile", "infile.jpeg", "filename of the image")
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "RobotoMono-SemiBold.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	fontsize     = flag.Float64("size", 12, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
	weekdays = []string{"Mo", "Di", "Mi", "Do", "Fr", "Sa", "So"}
)

func measureCell(drawer *font.Drawer, days []calendar.CalendarDay) image.Point {
	lineSize := int(math.Ceil(*fontsize * *spacing * *dpi / 72))
	maxLength := 0
	for _, day := range days {
		length := drawer.MeasureString(day.Format("2")).Round()
		if length > maxLength {
			maxLength = length
		}
	}
	for _, dayName := range weekdays {
		length := drawer.MeasureString(dayName).Round()
		if length > maxLength {
			maxLength = length
		}
	}
	spaceLength := drawer.MeasureString(" ").Round()
	maxLength += spaceLength
	return image.Pt(maxLength, lineSize)
}

func measureCalendar(drawer *font.Drawer, days []calendar.CalendarDay) image.Point {
	lineHeight := int(math.Ceil(*fontsize * *spacing * *dpi / 72))
	firstLineHeight := int(math.Ceil(*fontsize * *dpi / 72))
	cellSize := measureCell(drawer, days)
	numLines := len(days) / 7 + 2
	w := cellSize.X * 7
	h := firstLineHeight + (numLines - 1) * lineHeight
	return image.Pt(w, h)
}

func render(drawer *font.Drawer, days []calendar.CalendarDay, today time.Time, at image.Point) {
	lineHeight := int(math.Ceil(*fontsize * *spacing * *dpi / 72))
	firstLineHeight := int(math.Ceil(*fontsize * *dpi / 72))
	cellSize := measureCell(drawer, days)
	calSize := measureCalendar(drawer, days)

	x, y := at.X, at.Y
	dy := lineHeight
	dx := cellSize.X
	y += firstLineHeight
	headerText := today.Format("January 2006")
	headerLength := drawer.MeasureString(headerText).Round()
	drawer.Dot = fixed.P(x + int((calSize.X - headerLength) / 2), y)
	drawer.DrawString(headerText)
	y += dy
	x = at.X
	for _, dayName := range weekdays {
		length := drawer.MeasureString(dayName).Round()
		diff := cellSize.X - length
		drawer.Dot = fixed.P(x + diff, y)
		drawer.DrawString(dayName)
		x += dx
	}
	y += dy
	x = at.X
	for i, day := range days {
		if i != 0 && i % 7 == 0 {
			y += dy
			x = at.X
		}
		if day.ThisMonth() {
			text := day.Format("2")
			length := drawer.MeasureString(text).Round()
			diff := cellSize.X - length
			drawer.Dot = fixed.P(x + diff, y)
			drawer.DrawString(text)
		}
		x += dx
	}
}

func openImageFile(filename string) (image.Image, error) {
	inFile, err := os.Open(*imgfile)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()
	img, format, err := image.Decode(inFile)
	if err != nil {
		e := fmt.Errorf("error in decoding: %w", err)
		return nil, e
	}
	if format != "jpeg" && format != "png" {
		e := fmt.Errorf("error in image format - not jpeg")
                return nil, e
	}
	return img, nil
}

func openFontFile(filename string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		e := fmt.Errorf("error in parsing truetype: %w", err)
		return nil, e
	}
	return font, nil
}

func writeImageFile(filename string, rgba *image.RGBA) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	buf := bufio.NewWriter(outFile)
	err = png.Encode(buf, rgba)
	if err != nil {
		return fmt.Errorf("error in parsing truetype: %w", err)
	}
	err = buf.Flush()
	if err != nil {
		return err
	}
	return nil
}

func createDrawer(rgba *image.RGBA, tfont *truetype.Font, fg image.Image) *font.Drawer {
	h := font.HintingNone
	switch *hinting {
	case "full":
		h = font.HintingFull
	}
	return &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(tfont, &truetype.Options{
			Size:    *fontsize,
			DPI:     *dpi,
			Hinting: h,
		}),
	}
}

func main() {
	flag.Parse()

	// Read the font data.
	f, err := openFontFile(*fontfile)
	if err != nil {
		log.Fatal(err)
	}

	// Read the image
	img, err := openImageFile(*imgfile)
	if err != nil {
		log.Fatal(err)
	}

	// Copy the image data
	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rgba.Bounds(), img, rect.Min, draw.Src) 

	// Create the drawer
	fg, bg := image.Black, &image.Uniform{color.RGBA{255, 255, 255, 125}}
	if *wonb {
		fg, bg = image.White, &image.Uniform{color.RGBA{0, 0, 0, 125}}
	}
	drawer := createDrawer(rgba, f, fg)

	// Create calendar data
	today := time.Now()
	days := calendar.GetCalendar(today)
	calSize := measureCalendar(drawer, days)
	padding := image.Pt(20, 20)
	rr := image.Rectangle{ rect.Max.Sub(calSize).Sub(padding.Mul(2)), rect.Max.Sub(padding) }

	// Draw background
	draw.Draw(rgba, rr, bg, image.ZP, draw.Over)

	// Draw calendar
	render(drawer, days, today, rr.Min)

	// Save that RGBA image to disk.
	err = writeImageFile("out.png", rgba)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote out.png OK.")
}
