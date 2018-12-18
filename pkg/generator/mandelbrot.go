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

// Render generates the Mandelbrot set with N CPU number of go routines
func (m *Mandelbrot) Render() {
	m.RenderWithBufferedChannel()
}

// RenderSequentially generates the Mandelbrot set sequentially, without go routines
func (m *Mandelbrot) RenderSequentially() {
	m.ImageData = image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))

	nPixels := m.Height * m.Width

	for index := 0; index < nPixels*4; index += 4 {
		m.colorize(m.toCoordinate(index))
	}
}

// RenderWithUnlimitedRoutines uses one go routine per Coordinate
// All go routines read from one Coordinate channel with no buffer set
func (m *Mandelbrot) RenderWithUnlimitedRoutines() {
	m.ImageData = image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))

	nPixels := m.Height * m.Width

	done := make(chan bool, nPixels)

	for index := 0; index < nPixels*4; index += 4 {
		go func(c Coordinate) {
			m.colorize(c)
			done <- true
		}(m.toCoordinate(index))
	}

	for i := 0; i < nPixels; i++ {
		<-done
	}
}

// RenderWithMaxRoutines limits the number of go routines
// All go routines read from one Coordinate channel with no buffer set
func (m *Mandelbrot) RenderWithMaxRoutines(maxRoutines int) {
	m.ImageData = image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))

	nPixels := m.Height * m.Width

	semaphore := make(chan bool, maxRoutines)

	for index := 0; index < nPixels*4; index += 4 {
		semaphore <- true // occupy one slot
		go func(c Coordinate) {
			m.colorize(c)
			<-semaphore // free one slot
		}(m.toCoordinate(index))
	}

	// fill all slots to ensure remaining routines finish
	for i := 0; i < maxRoutines; i++ {
		semaphore <- true
	}
}

// RenderWithBufferedChannel uses a buffered channel for Coordinates
// The buffer size is tthe number of pixel devided by n CPU.
// For each CPU one go routine is started, each reading from the buffered
// Coorindate channel
func (m *Mandelbrot) RenderWithBufferedChannel() {
	m.ImageData = image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))
	nPixels := m.Height * m.Width

	coordinates := make(chan Coordinate, nPixels/4)
	go func() {
		for index := 0; index < nPixels*4; index += 4 {
			coordinates <- m.toCoordinate(index)
		}
		close(coordinates)
	}()

	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			defer wg.Done()
			for coordinate := range coordinates {
				m.colorize(coordinate)
			}
		}()
	}

	wg.Wait()
}

func (m *Mandelbrot) Coordinates(buffer int) chan Coordinate {
	m.ImageData = image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))
	nPixels := m.Height * m.Width

	coordinates := make(chan Coordinate, buffer)
	go func() {
		for index := 0; index < nPixels*4; index += 4 {
			coordinates <- m.toCoordinate(index)
		}
		close(coordinates)
	}()

	return coordinates
}

type MandelbrotResult struct {
	IsMandelbrot bool
	Iterations   int
	Re           float64
	Im           float64
	Index        int
}

func IsMandelbrot(coordinates chan Coordinate, resultChan chan MandelbrotResult, maxIterations int) {
	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			defer wg.Done()
			for coordinate := range coordinates {
				resultChan <- isMandelbrotToResult(coordinate, maxIterations)
			}
			fmt.Println("reading from coordinates channel Done")
		}()
	}

	defer close(resultChan)
	wg.Wait()
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

// toCoordinate maps an []RGBA index to a Coordinate in a complex plane
// []RGBA stores 4 numbers for each pixel (r, g, b, a) so every 4 indexes a new pixel starts
func (m *Mandelbrot) toCoordinate(index int) Coordinate {
	var width = float64(m.Width)
	var height = float64(m.Height)

	aspectRatio := height / width
	pixelIndex := int(math.Floor(float64(index) / 4))

	// Create a point in 0, 0 top and left indexed pane
	point := image.Point{pixelIndex % m.Width, int(math.Floor(float64(pixelIndex) / width))}

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

func isMandelbrot2(coordinate Coordinate, maxIterations int) (bool, int, float64, float64) {
	var cr = coordinate.Re
	var ci = coordinate.Im
	var zr = cr
	var zi = ci

	for iteration := 0; iteration < maxIterations; iteration++ {
		if zr*zr+zi*zi > 4 {
			return false, iteration, zr, zi
		}

		newzr := (zr * zr) - (zi * zi) + cr
		newzi := ((zr * zi) * 2) + ci
		zr = newzr
		zi = newzi
	}

	return true, maxIterations, zr, zi
}

func isMandelbrotToResult(coordinate Coordinate, maxIterations int) MandelbrotResult {
	isMandelbrot, iterations, re, im := isMandelbrot2(coordinate, maxIterations)
	result := MandelbrotResult{isMandelbrot, iterations, re, im, coordinate.Index}

	return result
}

// colorize takes a Coordinate and uses the results from isMandelbrot to
// get a color. The color is then set in Mandelbrot.ImageData
func (m *Mandelbrot) colorize(coordinate Coordinate) {
	isMandelbrot, iteration, real, imaginary := m.isMandelbrot(coordinate)
	color := colors.GetColor(m.Colors, isMandelbrot, iteration, m.MaxIterations, real, imaginary)
	m.ImageData.Pix[coordinate.Index] = color.R
	m.ImageData.Pix[coordinate.Index+1] = color.G
	m.ImageData.Pix[coordinate.Index+2] = color.B
	m.ImageData.Pix[coordinate.Index+3] = color.A
}

func (m *Mandelbrot) ColorizeFunc(isMandelbrot bool, iteration int, real float64, imaginary float64, maxIterations int, index int) {
	color := colors.GetColor(m.Colors, isMandelbrot, iteration, maxIterations, real, imaginary)
	m.ImageData.Pix[index] = color.R
	m.ImageData.Pix[index+1] = color.G
	m.ImageData.Pix[index+2] = color.B
	m.ImageData.Pix[index+3] = color.A
}
