# Mandelbrot

This Go library generates the Mandelbrot set and served as my playground to learn about concurrency in Go. Generating Mandelbrot set images sounded like a fun idea and a good fit because it requires a lot of calculations for a large amount of data. And you get some pretty images, Yay!
I wanted to find out how go routines and channels work and experimented with some concurrency approaches I came up with.

<p align="center">
    <img src="/assets/image01.jpeg" width="220">
    <img src="/assets/image02.jpeg" width="220">
    <img src="/assets/image03.jpeg" width="220">
    <img src="/assets/image04.jpeg" width="220">
</p>

## How to use it?

`mandelbrot` assumes you have Go 1.11+ installed as it uses Go modules for its dependencies.

```text
$ git clone https://github.com/saesh/mandelbrot && cd mandelbrot
$ go run examples/mandelbrot-jpeg/main.go
```

If you are using Go in version <1.11 you can install the depencencies manually:

```text
go get github.com/lucasb-eyer/go-colorful
```

---

The repository contains the library for generating Mandelbrot set image data in `pkg` and some example programs in `examples`.

The `Mandelbrot` type exposes the image data in the property `ImageData`, or it can be encoded as a JPEG file with the `WriteJpeg` method.

To create an image import the library and create a `Mandelbrot` object with parameters for the generation of the set:

```go
package main

import (
	"github.com/saesh/mandelbrot/pkg/colors"
	g "github.com/saesh/mandelbrot/pkg/generator"
)

func main() {
    mb := &g.Mandelbrot{}

    mb.Width = 1000
    mb.Height = 1000
    mb.MaxIterations = 600
    mb.Colors = colors.Hue

    mb.X = 0
    mb.Y = 0
    mb.R = 4

    mb.Render()
    mb.WriteJpeg("mandelbrot.jpeg", 90)
}
```

The `Render` method blocks until all pixels are generated. The quality of the JPEG file can be set with the second parameter to `WriteJpeg`.

## Concurrency

To understand go routines and channels I took several approaches to write this library in a concurrent fashion:

### Sequential rendering

`Mandelbrot.RenderSequentially`: Obviously sequential rendering is not concurrent but it was the first implementation of the algorithm and used a baseline. During this approach I found out, that `math.Pow` is incredibly
slow.

### Unlimited go routines

`Mandelbrot.RenderWithUnlimitedRoutines`: Next step was to spawn a go routine per Coordinate (pixel) and see what happens. The result were even longer render times than the sequential rendering. This is due to the coordination of millions of go routines in the go runtime. Although CPU usage was ~80% accross all cores the render times were awefully slow. Also memory usage went up as each go routine needs a certain amount of memory (~4KB). Having millions of them waiting lets the memory usage grow fast.

### Limit the maximum number of go routines

`Mandelbrot.RenderWithMaxRoutines`: As unlimited go routines were not really fast, the next approach was to limit the number of go routines. I limited them to 100. Again, each routine got one Coordinate to render. The render times were faster! But not by that much, still dissappointing.

### Buffered channel for N CPU go routines

`Mandelbrot.RenderWithBufferedChannel`: Next, the data was split up in batches. The number of batches is equal to the number of CPUs. And for each batch one go routine was spawned. The Coordinate channel's buffer size is that of the length of each batch. So on a 4 core system rendering a 4000 pixel image 4 go routines would spawn, each processing 1000 pixels. This was really fast and let all cores run at 100% with almost no system usage.

I plotted the render times of each approach:

<p align="center">
    <img src="/assets/rendering-times-chart.png">
</p>

The data was gathered on a 4 core Intel Core i5-3570 at 3.6Ghz, with Go 1.11.2 on a Linux system. The maximum number of iterations for the algorithm was 300.

## Credits

   - This article laid out the foundation for the Mandelbrot set calculations: https://blog.jfo.click/the-mandelwat-set/
   - Coloring the pixels based on iteration and complex numbers is hard, these answers on Stackoverflow helped immensely:
      - https://stackoverflow.com/a/16505538 
      - https://stackoverflow.com/a/1243788
      - https://stackoverflow.com/a/18678856
   - I didn't want to generate gradient colors myself, so I used `go-colorful` which includes a gradient generator in their docs: https://github.com/lucasb-eyer/go-colorful/blob/master/doc/gradientgen/gradientgen.go

## Further reading

   - https://blog.golang.org/pipelines
   - https://medium.com/@thejasbabu/concurrency-patterns-golang-5c5e1bcd0833
