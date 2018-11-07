package colors

import "image/color"

var palette [16]color.RGBA

func init() {
	palette[0] = color.RGBA{66, 30, 15, 255}
	palette[1] = color.RGBA{25, 7, 26, 255}
	palette[2] = color.RGBA{9, 1, 47, 255}
	palette[3] = color.RGBA{4, 4, 73, 255}
	palette[4] = color.RGBA{0, 7, 100, 255}
	palette[5] = color.RGBA{12, 44, 138, 255}
	palette[6] = color.RGBA{24, 82, 177, 255}
	palette[7] = color.RGBA{57, 125, 209, 255}
	palette[8] = color.RGBA{134, 181, 229, 255}
	palette[9] = color.RGBA{211, 236, 248, 255}
	palette[10] = color.RGBA{241, 233, 191, 255}
	palette[11] = color.RGBA{248, 201, 95, 255}
	palette[12] = color.RGBA{255, 170, 0, 255}
	palette[13] = color.RGBA{204, 128, 0, 255}
	palette[14] = color.RGBA{153, 87, 0, 255}
	palette[15] = color.RGBA{106, 52, 3, 255}
}

func fromPalette(iteration int) color.RGBA {
	return palette[iteration%len(palette)]
}
