package main

import (
	"image"
	"image/draw"
	"log"
	"sort"

	"github.com/fogleman/gg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	inputFile   = kingpin.Flag("input", "Input image.").Required().Short('i').ExistingFile()
	outputFile  = kingpin.Flag("output", "Output image.").Required().Short('o').String()
	windowSize  = kingpin.Flag("size", "Window size as a percentage.").Short('s').Default("0.05").Float64()
	percentile  = kingpin.Flag("percentile", "Window percentile.").Short('p').Default("0.5").Float64()
	targetValue = kingpin.Flag("target", "Target value when scaling output.").Short('t').Default("240").Int()
)

func ensureGray(im image.Image) (*image.Gray, bool) {
	switch im := im.(type) {
	case *image.Gray:
		return im, true
	default:
		dst := image.NewGray(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst, false
	}
}

func windowPercentile(im *image.Gray, r image.Rectangle, p float64) float64 {
	var values []float64
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		for x := r.Min.X; x <= r.Max.X; x++ {
			values = append(values, float64(im.GrayAt(x, y).Y))
		}
	}
	sort.Float64s(values)
	i := int(float64(len(values))*p+0.5) - 1
	return values[i]
}

func main() {
	kingpin.Parse()

	d := 0
	s := *windowSize
	p := *percentile
	t := float64(*targetValue)

	src, err := gg.LoadImage(*inputFile)
	if err != nil {
		log.Fatal(err)
	}
	im, _ := ensureGray(src)
	w := im.Bounds().Size().X
	h := im.Bounds().Size().Y
	sx := int(float64(w)*s + 0.5)
	sy := int(float64(h)*s + 0.5)
	dx := w - (d+sx)*2
	dy := h - (d+sy)*2
	r00 := image.Rect(d, d, d+sx, d+sy)
	r10 := r00.Add(image.Pt(dx, 0))
	r01 := r00.Add(image.Pt(0, dy))
	r11 := r00.Add(image.Pt(dx, dy))
	a00 := windowPercentile(im, r00, p)
	a01 := windowPercentile(im, r01, p)
	a10 := windowPercentile(im, r10, p)
	a11 := windowPercentile(im, r11, p)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := float64(x) / float64(w-1)
			py := float64(y) / float64(h-1)
			ax0 := a00*(1-px) + a10*px
			ax1 := a01*(1-px) + a11*px
			a := ax0*(1-py) + ax1*py
			v := float64(im.Pix[i])
			v = (v / a) * t
			if v < 0 {
				v = 0
			}
			if v > 255 {
				v = 255
			}
			im.Pix[i] = uint8(v)
			i++
		}
	}
	err = gg.SavePNG(*outputFile, im)
	if err != nil {
		log.Fatal(err)
	}
}
