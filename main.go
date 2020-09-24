package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	colors "golang.org/x/image/colornames"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/tools/bezier"
	"gonum.org/v1/plot/vg"
)

const (
	screenWidth              = 1280
	screenHeight             = 720
	offsetX          float64 = 400
	offsetY          float64 = 300
	scaleX           float64 = 300
	scaleY           float64 = 300
	numberOfSegments         = 10
	epsilon          float64 = 0.001
	dt               float64 = 0.5
)

func gonumToPixel(xy plotter.XY) pixel.Vec {
	return pixel.V(xy.X, xy.Y)
}

func getSegmentPoints(points plotter.XYs, numberOfSegments int) []pixel.Vec {
	// Create lines out of bezier
	// curve points.
	lines := []pixel.Line{}

	for i := 0; i < len(points)-1; i++ {
		line := pixel.L(gonumToPixel(points[i]),
			gonumToPixel(points[i+1]))

		lines = append(lines, line)
	}

	// Compute the length
	// of the bezier curve
	// interpolated with lines.
	length := 0.0

	for _, line := range lines {
		length += line.Len()
	}

	// Divide the bezier curve into
	// equal segments.
	step := length / float64(numberOfSegments)
	segmentPoints := []pixel.Vec{}
	lastLine := 0
	lastPoint := lines[0].A
	segmentPoints = append(segmentPoints, lastPoint)

	for i := 0; i < numberOfSegments; i++ {
		subsegments := []pixel.Line{}
		startLine := pixel.L(lastPoint, lines[lastLine].B)

		subsegments = append(subsegments, startLine)
		localLength := startLine.Len()

		for step-localLength > epsilon {
			line := lines[lastLine+1]
			subsegments = append(subsegments, line)

			localLength += line.Len()
			lastLine++
		}

		line := lines[lastLine]

		if localLength-step > epsilon {
			difference := localLength - step
			t := difference / line.Len()

			lastPoint = pixel.V(t*line.A.X+(1-t)*line.B.X,
				t*line.A.Y+(1-t)*line.B.Y)
		} else {
			lastPoint = line.B
			lastLine++
		}

		segmentPoints = append(segmentPoints, lastPoint)
	}

	return segmentPoints
}

func run() {
	controlPoints := []vg.Point{
		{X: 0.45, Y: 0.328},
		{X: 1.403, Y: 0.12},
		{X: 0.62, Y: 1.255},
		{X: 1.521, Y: 0.593},
	}

	// Form the curve.
	curve := bezier.New(controlPoints...)
	points := make(plotter.XYs, 0)

	for t := 0.0; t < 100.0; t += dt {
		point := curve.Point(t / 100.0)

		points = append(points, plotter.XY{
			X: float64(point.X)*scaleX + offsetX,
			Y: float64(point.Y)*scaleY + offsetY})
	}

	// Divide the curve into equal segments.
	segmentPoints := getSegmentPoints(points, numberOfSegments)

	cfg := pixelgl.WindowConfig{
		Title:  "Bezier curve",
		Bounds: pixel.R(0, 0, screenWidth, screenHeight),
	}
	win, err := pixelgl.NewWindow(cfg)
	handleError(err)

	imd := imdraw.New(nil)

	fps := 0
	perSecond := time.Tick(time.Second)

	for !win.Closed() {
		win.Clear(colors.White)
		imd.Clear()

		// Draw the curve and other things.
		/*imd.Color = colors.Red

		for _, point := range points {
			imd.Push(gonumToPixel(point))
			imd.Circle(1, 1)
		}*/

		// Draw the control points.
		imd.Color = colors.Blue

		for _, point := range controlPoints {
			imd.Push(pixel.V(float64(point.X), float64(point.Y)))
			imd.Circle(5, 0)
		}

		// Draw the curve segments.
		imd.Color = colors.Green

		for _, point := range segmentPoints {
			imd.Push(point)
			imd.Circle(3, 1)
		}

		imd.Draw(win)

		win.Update()

		// Show FPS in the window title.
		fps++

		select {
		case <-perSecond:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, fps))
			fps = 0

		default:
		}
	}
}

func main() {
	pixelgl.Run(run)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
