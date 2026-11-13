// Command brandgen renders the asbuilt brand assets.
//
// The mark is drawn in code so that the raster files can always be rebuilt
// from source, and so that the geometry stays identical between the SVG and
// the PNG versions.
//
// Usage:
//
//	go run ./tools/brandgen -out assets/brand
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	out := flag.String("out", "assets/brand", "directory to write the PNG files to")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "brandgen:", err)
		os.Exit(1)
	}

	for _, size := range []int{512, 64} {
		img := drawMark(size)
		path := filepath.Join(*out, fmt.Sprintf("logo-%d.png", size))
		if err := writePNG(path, img); err != nil {
			fmt.Fprintln(os.Stderr, "brandgen:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", path)
	}
}

// drawMark renders the mark: a die outline with four traces, the last of
// which is dashed the way an as-built drawing annotates a deviation from
// the design intent.
func drawMark(size int) *image.RGBA {
	const units = 64.0
	scale := float64(size) / units
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	ink := color.RGBA{0x10, 0x14, 0x1c, 0xff}
	width := int(math.Round(3.5 * scale))
	if width < 1 {
		width = 1
	}

	// Die outline, rounded corners approximated by quarter circles.
	drawRoundedRect(img, int(7*scale), int(7*scale), int(57*scale), int(57*scale), int(9*scale), ink, width)

	// Three solid traces.
	for _, y := range []int{20, 29, 38} {
		drawLine(img, int(18*scale), int(float64(y)*scale), int(46*scale), int(float64(y)*scale), ink, width)
	}

	// The fourth trace is dashed.
	y := 47 * scale
	for x := 18.0; x < 46; {
		x2 := math.Min(x+7, 46)
		drawLine(img, int(x*scale), int(y), int(x2*scale), int(y), ink, width)
		x += 13
	}

	return img
}

// drawLine draws a horizontal line with the given thickness.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, width int) {
	drawRect(img, x0, y0-width/2, x1-x0+1, width, c)
}

// drawRect fills a rectangle.
func drawRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

// drawRoundedRect draws a rectangle with rounded corners.
func drawRoundedRect(img *image.RGBA, x0, y0, x1, y1, radius int, c color.RGBA, width int) {
	// Straight edges.
	drawRect(img, x0+radius, y0, x1-x0-2*radius, width, c)
	drawRect(img, x0+radius, y1-width, x1-x0-2*radius, width, c)
	drawRect(img, x0, y0+radius, width, y1-y0-2*radius, c)
	drawRect(img, x1-width, y0+radius, width, y1-y0-2*radius, c)

	// Corners: an arc of radius "radius" with a stroke of the given width.
	corners := []struct{ cx, cy, sx, sy int }{
		{x0 + radius, y0 + radius, -1, -1},
		{x1 - radius, y0 + radius, 1, -1},
		{x0 + radius, y1 - radius, -1, 1},
		{x1 - radius, y1 - radius, 1, 1},
	}
	for _, corner := range corners {
		for dy := 0; dy <= radius; dy++ {
			for dx := 0; dx <= radius; dx++ {
				d := math.Hypot(float64(dx), float64(dy))
				if d > float64(radius) || d < float64(radius-width) {
					continue
				}
				img.Set(corner.cx+corner.sx*dx, corner.cy+corner.sy*dy, c)
			}
		}
	}
}

// writePNG writes the image as a PNG file.
func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
