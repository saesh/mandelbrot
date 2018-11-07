package colors

import (
	"image/color"
)

const BitShiftRgb = 1
const StaticPalette = 2
const Hue = 3
const GradientUltraFractal = 4

var black = color.RGBA{0, 0, 0, 255}

// GetColor ...
func GetColor(theme int, isMandelbrot bool, iteration int, maxIterations int, real float64, imaginary float64) color.RGBA {

	if isMandelbrot {
		return black
	}

	switch theme {
	case BitShiftRgb:
		return bitShiftRgb(iteration, maxIterations)
	case StaticPalette:
		return fromPalette(iteration)
	case Hue:
		return hue(iteration, real, imaginary)
	case GradientUltraFractal:
		return gradient(ultraFractalGradientColors, iteration, real, imaginary)
	default:
		return gradient(ultraFractalGradientColors, iteration, real, imaginary)
	}
}
