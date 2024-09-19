package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/alexa-infra/kalendar/internal/calendar"
	. "github.com/alexa-infra/kalendar/internal/context"
)

func measureCell(cfg config, drawer *font.Drawer, days []calendar.CalendarDay) image.Point {
	lineSize := cfg.lineHeight()
	maxLength := 0
	for _, day := range days {
		length := drawer.MeasureString(day.Format("2")).Round()
		if length > maxLength {
			maxLength = length
		}
	}
	for _, dayName := range calendar.Weekdays {
		length := drawer.MeasureString(dayName).Round()
		if length > maxLength {
			maxLength = length
		}
	}
	spaceLength := drawer.MeasureString(" ").Round()
	maxLength += spaceLength
	return image.Pt(maxLength, lineSize)
}

func measureCalendar(cfg config, drawer *font.Drawer, days []calendar.CalendarDay) image.Point {
	lineHeight := cfg.lineHeight()
	firstLineHeight := cfg.lineHeightNoSpacing()
	cellSize := measureCell(cfg, drawer, days)
	numLines := len(days)/7 + 2
	w := cellSize.X * 7
	h := firstLineHeight + (numLines-1)*lineHeight
	return image.Pt(w, h)
}

func (cfg config) lineHeight() int {
	return int(math.Ceil(cfg.fontsize * cfg.spacing * cfg.dpi / 72))
}

func (cfg config) lineHeightNoSpacing() int {
	return int(math.Ceil(cfg.fontsize * cfg.dpi / 72))
}

func render(cfg config, drawer *font.Drawer, days []calendar.CalendarDay, today time.Time, at image.Point) {
	lineHeight := cfg.lineHeight()
	firstLineHeight := cfg.lineHeightNoSpacing()
	cellSize := measureCell(cfg, drawer, days)
	calSize := measureCalendar(cfg, drawer, days)

	x, y := at.X, at.Y
	dy := lineHeight
	dx := cellSize.X
	y += firstLineHeight
	headerText := today.Format("January 2006")
	headerLength := drawer.MeasureString(headerText).Round()
	drawer.Dot = fixed.P(x+int((calSize.X-headerLength)/2), y)
	drawer.DrawString(headerText)
	y += dy
	x = at.X
	for _, dayName := range calendar.Weekdays {
		length := drawer.MeasureString(dayName).Round()
		diff := cellSize.X - length
		drawer.Dot = fixed.P(x+diff, y)
		drawer.DrawString(dayName)
		x += dx
	}
	y += dy
	x = at.X
	for i, day := range days {
		if i != 0 && i%7 == 0 {
			y += dy
			x = at.X
		}
		if day.ThisMonth() {
			text := day.Format("2")
			length := drawer.MeasureString(text).Round()
			diff := cellSize.X - length
			drawer.Dot = fixed.P(x+diff, y)
			drawer.DrawString(text)
		}
		x += dx
	}
}

func openImageFile(filename string) (image.Image, error) {
	inFile, err := os.Open(filename)
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
		return fmt.Errorf("error in encoding png: %w", err)
	}
	err = buf.Flush()
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %s OK.\n", filename)
	return nil
}

func createDrawer(cfg config, rgba *image.RGBA, tfont *truetype.Font, fg image.Image) *font.Drawer {
	h := font.HintingNone
	switch cfg.hinting {
	case "full":
		h = font.HintingFull
	}
	return &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(tfont, &truetype.Options{
			Size:    cfg.fontsize,
			DPI:     cfg.dpi,
			Hinting: h,
		}),
	}
}

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

type config struct {
	imgfile  string
	dpi      float64
	fontfile string
	hinting  string
	fontsize float64
	spacing  float64
	wonb     bool
	outfile  string
}

func parseConfig(ctx context.Context) (config, error) {
	stderr := GetContextStderr(ctx)
	args := GetContextArgs(ctx)

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	imgfile := flagSet.String("imgfile", "infile.jpeg", "filename of the image")
	outfile := flagSet.String("outfile", "out.png", "output filename")
	dpi := flagSet.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile := flagSet.String("fontfile", "RobotoMono-SemiBold.ttf", "filename of the ttf font")
	hinting := flagSet.String("hinting", "none", "none | full")
	fontsize := flagSet.Float64("size", 12, "font size in points")
	spacing := flagSet.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb := flagSet.Bool("whiteonblack", false, "white text on a black background")
	flagSet.SetOutput(stderr)
	err := flagSet.Parse(args[1:])
	if err != nil {
		return config{}, err
	}
	return config{*imgfile, *dpi, *fontfile, *hinting, *fontsize, *spacing, *wonb, *outfile}, nil
}

func run(ctx context.Context) error {
	cfg, err := parseConfig(ctx)
	if err != nil {
		return err
	}

	// Read the font data.
	f, err := openFontFile(cfg.fontfile)
	if err != nil {
		return err
	}

	// Read the image
	img, err := openImageFile(cfg.imgfile)
	if err != nil {
		return err
	}

	// Copy the image data
	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rgba.Bounds(), img, rect.Min, draw.Src)

	// Create the drawer
	fg, bg := image.Black, &image.Uniform{color.RGBA{255, 255, 255, 125}}
	if cfg.wonb {
		fg, bg = image.White, &image.Uniform{color.RGBA{0, 0, 0, 125}}
	}
	drawer := createDrawer(cfg, rgba, f, fg)

	// Create calendar data
	today := time.Now()
	days := calendar.GetCalendar(today)
	calSize := measureCalendar(cfg, drawer, days)
	padding := image.Pt(20, 20)
	rr := image.Rectangle{rect.Max.Sub(calSize).Sub(padding.Mul(2)), rect.Max.Sub(padding)}

	// Draw background
	draw.Draw(rgba, rr, bg, image.ZP, draw.Over)

	// Draw calendar
	render(cfg, drawer, days, today, rr.Min)

	// Save that RGBA image to disk.
	return writeImageFile(cfg.outfile, rgba)
}
