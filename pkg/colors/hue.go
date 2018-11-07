package colors

import (
	"image/color"
	"math"

	colorful "github.com/lucasb-eyer/go-colorful"
)

func hue(iteration int, real float64, imaginary float64) color.RGBA {
	zn := math.Sqrt(real*real + imaginary*imaginary)

	hue := float64(iteration) + 1 - math.Log(math.Log(math.Abs(zn)))/math.Log(2)

	if hue > 360 {
		hue -= 360
	}

	if hue < 0 {
		hue += 360
	}

	c := colorful.Hsv(hue, 1.0, 1.0)

	return color.RGBA{uint8(c.R * 255), uint8(c.G * 255), uint8(c.B * 255), 255}
}
