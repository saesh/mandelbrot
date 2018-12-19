package generator

import (
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/saesh/mandelbrot/pkg/colors"
)

// Coordinate holds real number 'Re', imaginary number 'Im' and corresponding pixel index
type Coordinate struct {
	Re    float64
	Im    float64
	Index int
}

// Mandelbrot defines the parameters of the set to render
// 	Width & Height define the dimensions of the output image
//	R specifies the "zoom"
//	MaxIterations is upper limit of iterations for the calculation if a pixel is contained in the set or not
//	X & Y define where the center of the complex plane should be from -2..2
type Mandelbrot struct {
	Width         int
	Height        int
	R             float64
	MaxIterations int
	X             float64
	Y             float64
	ImageData     *image.RGBA
	Colors        int
}

type MandelbrotResult struct {
	IsMandelbrot bool
	Iterations   int
	Re           float64
	Im           float64
	Index        int
}

func NewMandelbrot(width int, height int) *Mandelbrot {
	return &Mandelbrot{
		Width:         width,
		Height:        height,
		X:             0,
		Y:             0,
		R:             4,
		MaxIterations: 300,
		Colors:        colors.GradientUltraFractal,
		ImageData:     image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

// Render generates the Mandelbrot set with N CPU number of go routines
func (m *Mandelbrot) Render() {
	numCPU := runtime.NumCPU()
	buffer := m.Height * m.Width / numCPU
	coordinates := m.Coordinates(buffer)
	var wg sync.WaitGroup
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			defer wg.Done()
			for coordinate := range coordinates {
				isMandelbrot, it, r, im := m.isMandelbrot(coordinate)
				m.ColorizeFunc(isMandelbrot, it, r, im, m.MaxIterations, coordinate.Index)
			}
		}()
	}

	wg.Wait()
}

func (m *Mandelbrot) Coordinates(buffer int) chan Coordinate {
	nPixels := m.Height * m.Width

	coordinates := make(chan Coordinate, buffer)
	go func() {
		for index := 0; index < nPixels; index++ {
			coordinates <- m.toCoordinate(index)
		}
		close(coordinates)
	}()

	return coordinates
}

func (m *Mandelbrot) IsMandelbrot(coordinates chan Coordinate, resultChan chan MandelbrotResult) {
	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			defer wg.Done()
			for coordinate := range coordinates {
				resultChan <- m.isMandelbrotToResult(coordinate)
			}
		}()
	}

	defer close(resultChan)
	wg.Wait()
}

// toCoordinate maps an []RGBA index to a Coordinate in a complex plane
// []RGBA stores 4 numbers for each pixel (r, g, b, a) so every 4 indexes a new pixel starts
func (m *Mandelbrot) toCoordinate(index int) Coordinate {
	var width = float64(m.Width)
	var height = float64(m.Height)

	aspectRatio := height / width

	// Create a point in 0, 0 top and left indexed pane
	point := image.Point{index % m.Width, int(math.Floor(float64(index) / width))}

	// Move pane to a complex pane, 0,0 centered and scale with factor
	re := (((float64(point.X) * m.R / width) - m.R/2) + (m.X * aspectRatio)) / aspectRatio
	im := (((float64(point.Y) * m.R / height) - m.R/2) * -1) + m.Y

	return Coordinate{re, im, index}
}

// isMandelbrot calculates if a Cooridinate on the complex plane is within the Mandelbrot set or not
// returns bool wether the Coorindate is in the Mandelbrot set
//         int iteration number when the algorithm exited
//         float64 the real number calculated up to the point when the algorithm exited
//         float64 the imaginary number calculated up to the point when the algorithm exited
// The returned values are used to calculate the color of the corresponding pixel
func (m *Mandelbrot) isMandelbrot(coordinate Coordinate) (bool, int, float64, float64) {
	var cr = coordinate.Re
	var ci = coordinate.Im
	var zr = cr
	var zi = ci

	for iteration := 0; iteration < m.MaxIterations; iteration++ {
		if zr*zr+zi*zi > 4 {
			return false, iteration, zr, zi
		}

		newzr := (zr * zr) - (zi * zi) + cr
		newzi := ((zr * zi) * 2) + ci
		zr = newzr
		zi = newzi
	}

	return true, m.MaxIterations, zr, zi
}

func (m *Mandelbrot) isMandelbrotToResult(coordinate Coordinate) MandelbrotResult {
	isMandelbrot, iterations, re, im := m.isMandelbrot(coordinate)
	result := MandelbrotResult{isMandelbrot, iterations, re, im, coordinate.Index}

	return result
}

func (m *Mandelbrot) ColorizeFunc(isMandelbrot bool, iteration int, real float64, imaginary float64, maxIterations int, index int) {
	color := colors.GetColor(m.Colors, isMandelbrot, iteration, maxIterations, real, imaginary)
	arrayIndex := index * 4
	m.ImageData.Pix[arrayIndex] = color.R
	m.ImageData.Pix[arrayIndex+1] = color.G
	m.ImageData.Pix[arrayIndex+2] = color.B
	m.ImageData.Pix[arrayIndex+3] = color.A
}

// WriteJpeg encodes the Mandelbrot.ImageData to a JPEG file
// The quality for JPEG encoding must be passed as int, value between 0..100
func (m *Mandelbrot) WriteJpeg(filename string, quality int) error {
	outputFile, err := os.Create(filename)
	defer outputFile.Close()

	if err != nil {
		fmt.Println("could not create file.")
		return err
	}

	options := jpeg.Options{Quality: quality}
	jpeg.Encode(outputFile, m.ImageData, &options)

	return nil
}
