package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/saesh/mandelbrot/pkg/colors"
	g "github.com/saesh/mandelbrot/pkg/generator"
)

const (
	renderStatsFilename = "rendering-times.dat"
	maxResolution       = 3000
)

func main() {
	mb := &g.Mandelbrot{
		MaxIterations: 300,
		Colors:        colors.GradientUltraFractal,
		X:             0,
		Y:             0,
		R:             4,
	}

	deleteFile(renderStatsFilename)

	for length := 100; length <= maxResolution; length += 100 {
		mb.Width = length
		mb.Height = length

		col1 := fmt.Sprintf("%vx%v", mb.Width, mb.Height)
		col2 := measureRenderingTime(mb.RenderSequentially, *mb)
		col3 := measureRenderingTime(mb.RenderWithUnlimitedRoutines, *mb)
		col4 := measureRenderingTime(func() { mb.RenderWithMaxRoutines(100) }, *mb)
		col5 := measureRenderingTime(mb.RenderWithBufferedChannel, *mb)

		appendToFile(renderStatsFilename, fmt.Sprintf("%s   %v   %v   %v   %v\n", col1, col2, col3, col4, col5))
	}
}

func measureRenderingTime(fn func(), mb g.Mandelbrot) float64 {
	var start = time.Now()

	fn()
	elapsed := time.Since(start).Seconds()

	mb.ImageData = nil
	runtime.GC()

	return elapsed
}

func deleteFile(filename string) {
	os.Remove(filename)
}

func appendToFile(filename string, line string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(line); err != nil {
		panic(err)
	}
}
