package main

import (
	"log"
	"time"

	"github.com/saesh/mandelbrot/pkg/colors"
	g "github.com/saesh/mandelbrot/pkg/generator"
)

func main() {
	mb := g.NewMandelbrot(6000, 6000)

	mb.MaxIterations = 300
	mb.Colors = colors.GradientUltraFractal

	mb.X = 0
	mb.Y = 0
	mb.R = 4

	// mb.X = -0.74453985651
	// mb.Y = 0.12172277365
	// mb.R = 0.000003072

	// mb.X = -0.74515
	// mb.Y = 0.11245
	// mb.R = 0.00065

	// mb.X = -0.744297086329353
	// mb.Y = 0.15142492223558
	// mb.R = 0.016

	// mb.X = -0.748
	// mb.Y = 0.1
	// mb.R = 0.0025

	start := time.Now()
	mb.Render()
	log.Printf("elapsed: %v\n", time.Since(start))
	mb.WriteJpeg("mandelbrot.jpeg", 90)
}
