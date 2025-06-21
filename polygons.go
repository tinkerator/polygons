// Program polygons demostrates the determination of a
// non-intersecting outline of a collection of polygons. It outlines
// solids and voids ("Holes") in different colors.
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"zappem.net/pub/graphics/raster"
	"zappem.net/pub/math/polygon"
)

func pt(x, y float64) polygon.Point {
	return polygon.Point{X: x, Y: y}
}

func triangle(p *polygon.Shapes, x, y, w float64) *polygon.Shapes {
	return p.Builder(pt(x, y), pt(x+2*w, y), pt(x+w, y+4/3*w))
}

func untriangle(p *polygon.Shapes, x, y, w float64) *polygon.Shapes {
	return p.Builder(pt(x, y), pt(x+w, y+4/3*w), pt(x+2*w, y))
}

type Affine struct {
	x0, m, y0 float64
}

func NewAffine(from0, from, to0, to float64) *Affine {
	return &Affine{
		x0: from0,
		y0: to0,
		m:  (to - to0) / (from - from0),
	}
}

func (a *Affine) To(x float64) float64 {
	return (a.m*(x-a.x0) + a.y0)
}

func visualize(name string, before, after *polygon.Shapes) {
	im := image.NewRGBA(image.Rect(0, 0, 500, 250))
	draw.Draw(im, im.Bounds(), &image.Uniform{color.RGBA{0xff, 0xff, 0xff, 0xff}}, image.ZP, draw.Src)
	rast := raster.NewRasterizer()
	mB, xB := 0., 0.
	started := false
	for _, p := range before.P {
		if mB > p.MinX || mB > p.MinY || !started {
			mB, _ = polygon.MinMax(p.MinX, p.MinY)
		}
		if xB < p.MaxX || xB < p.MaxY || !started {
			_, xB = polygon.MinMax(p.MaxX, p.MaxY)
		}
		started = true
	}
	bAx := NewAffine(mB, xB, 10, 240)
	bAy := NewAffine(mB, xB, 240, 10)
	mA, xA := 0., 0.
	started = false
	for _, p := range after.P {
		if mA > p.MinX || mA > p.MinY || !started {
			mA, _ = polygon.MinMax(p.MinX, p.MinY)
		}
		if xA < p.MaxX || xA < p.MaxY || !started {
			_, xA = polygon.MinMax(p.MaxX, p.MaxY)
		}
		started = true
	}
	aAx := NewAffine(mA, xA, 260, 490)
	aAy := NewAffine(mA, xA, 240, 10)

	for _, pts := range before.P {
		for j := 0; j < len(pts.PS); j++ {
			was := pts.PS[j]
			is := pts.PS[(j+1)%len(pts.PS)]
			fromX, fromY := bAx.To(was.X), bAy.To(was.Y)
			toX, toY := bAx.To(is.X), bAy.To(is.Y)
			raster.LineTo(rast, true, fromX, fromY, toX, toY, 3)
		}
		col := color.RGBA{0xff, 0x00, 0x00, 0xff}
		if pts.Hole {
			col = color.RGBA{0x00, 0x00, 0xff, 0xff}
		}
		rast.Render(im, 0, 0, col)
		rast.Reset()
	}
	for _, pts := range before.P {
		col := color.RGBA{0x00, 0x00, 0x00, 0xff}
		for j := 0; j < len(pts.PS); j++ {
			pt := pts.PS[j]
			x, y := bAx.To(pt.X), bAy.To(pt.Y)
			raster.PointAt(rast, x, y, 4)
			rast.Render(im, 0, 0, col)
			rast.Reset()
		}
	}
	for _, pts := range after.P {
		for j := 0; j < len(pts.PS); j++ {
			was := pts.PS[j]
			is := pts.PS[(j+1)%len(pts.PS)]
			fromX, fromY := aAx.To(was.X), aAy.To(was.Y)
			toX, toY := aAx.To(is.X), aAy.To(is.Y)
			raster.LineTo(rast, true, fromX, fromY, toX, toY, 3)
		}
		col := color.RGBA{0xff, 0x00, 0x00, 0xff}
		if pts.Hole {
			col = color.RGBA{0x00, 0x00, 0xff, 0xff}
		}
		rast.Render(im, 0, 0, col)
		rast.Reset()
	}
	for _, pts := range after.P {
		col := color.RGBA{0x00, 0x00, 0x00, 0xff}
		for j := 0; j < len(pts.PS); j++ {
			pt := pts.PS[j]
			x, y := aAx.To(pt.X), aAy.To(pt.Y)
			raster.PointAt(rast, x, y, 4)
			rast.Render(im, 0, 0, col)
			rast.Reset()
		}
	}

	// generate image for before (determine its height and width)
	// generate image for after (extend this height and width)
	// create an RGBA image big enough
	// write it to file

	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("failed to create %q: %v", name, err)
	}
	defer f.Close()
	png.Encode(f, im)
	log.Printf("wrote result to %q", name)
}

func main() {
	ps := triangle(nil, 9, 10, 4)
	ps = triangle(ps, 13, 12, 4)
	ps = untriangle(ps, 11, 11, 1)
	ps = triangle(ps, 13, 9, 4)
	ps = triangle(ps, 20, 15, 4)
	dup := ps.Duplicate()
	ps.Union()
	visualize("dump.png", dup, ps)
}
