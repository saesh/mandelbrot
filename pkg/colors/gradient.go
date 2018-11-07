package colors

import (
	"image/color"
	"math"

	colorful "github.com/lucasb-eyer/go-colorful"
)

var ultraFractalGradientColors [2048]colorful.Color

func init() {
	generateUltraFractalGradient()
}

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func (g GradientTable) getInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(g)-1; i++ {
		c1 := g[i]
		c2 := g[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendRgb(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return g[len(g)-1].Col
}

func generateUltraFractalGradient() {
	c1, _ := colorful.MakeColor(color.RGBA{0, 7, 100, 255})
	c2, _ := colorful.MakeColor(color.RGBA{32, 107, 203, 255})
	c3, _ := colorful.MakeColor(color.RGBA{237, 255, 255, 255})
	c4, _ := colorful.MakeColor(color.RGBA{255, 170, 0, 255})
	c5, _ := colorful.MakeColor(color.RGBA{0, 2, 0, 255})
	c6, _ := colorful.MakeColor(color.RGBA{0, 7, 100, 255})

	keypoints := GradientTable{
		{c1, 0.0},
		{c2, 0.16},
		{c3, 0.42},
		{c4, 0.6425},
		{c5, 0.8575},
		{c6, 1.0},
	}

	for y := 2047; y >= 0; y-- {
		ultraFractalGradientColors[2047-y] = keypoints.getInterpolatedColorFor(float64(y) / 2048)
	}
}

func gradient(gradientColors [2048]colorful.Color, iteration int, real float64, imaginary float64) color.RGBA {
	const OneOverLog2 = 1.4426950408889634
	const GradientScale = 256
	const GradientShift = 0

	size := math.Sqrt(real*real + imaginary*imaginary)
	smoothed := math.Log(math.Log(size)*OneOverLog2) * OneOverLog2
	colorI := int(math.Sqrt(float64(iteration+1)-smoothed)*GradientScale+GradientShift) % 2048

	r, g, b := gradientColors[colorI].RGB255()

	return color.RGBA{r, g, b, 255}
}
