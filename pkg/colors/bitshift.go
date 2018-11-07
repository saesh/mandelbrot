package colors

import "image/color"

func bitShiftRgb(iteration int, maxIterations int) color.RGBA {
	// normalize iteration to be between 0 and 1
	iterationNormalized := float64(iteration) / float64(maxIterations)

	// scale iteration to be between 0 and 16777215
	iterationScaled := iterationNormalized * 0xffffff

	// extract r, g, b values by bitshifting
	r := uint8(uint32(iterationScaled) & 0xff0000 >> 16)
	g := uint8(uint32(iterationScaled) & 0x00ff00 >> 8)
	b := uint8(uint32(iterationScaled) & 0x0000ff)

	return color.RGBA{r, g, b, 255}
}
