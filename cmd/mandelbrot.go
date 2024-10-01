package main

import (
	"fmt"
	"image/png"
	"math"
	"os"

	"github.com/fogleman/gg"
	"github.com/greglange/mandelbrot/pkg/mandelbrot"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: mandelbrot config.yml")
		return
	}
	var c mandelbrot.Config
	err := cleanenv.ReadConfig(os.Args[1], &c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else if c.RunType == "image" {
		err := mainImage(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else if c.RunType == "video" {
		err := mainVideo(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else if c.RunType == "test" {
		err := mainTest(c)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Invalid run type")
	}
}

func mainTest(config mandelbrot.Config) error {
	type corner struct {
		x int
		y int
	}
	corners := []corner{
		corner{0, 0},
		corner{config.ImageWidth, 0},
		corner{0, config.ImageHeight},
		corner{config.ImageWidth, config.ImageHeight},
	}
	for _, c := range corners {
		a, b := mandelbrot.ImageToPlane(config, c.x, c.y)
		fmt.Printf("%d,%d -> %f,%f\n", c.x, c.y, a, b)
	}
	return nil
}

func mainImage(c mandelbrot.Config) error {
	var err error
	calcColor, err := mandelbrot.GetCalculateColor(c)
	if err != nil {
		return err
	}
	dc, err := mandelbrot.GetImage(c, c.CenterX, c.CenterY, c.Scale, c.Rotation, calcColor)
	if err != nil {
		return err
	}
	err = mandelbrot.DrawLines(c, dc)
	if err != nil {
		return err
	}
	png.Encode(os.Stdout, dc.Image())
	return nil
}

func mainVideo(c mandelbrot.Config) error {
	var (
		dc  *gg.Context
		err error
	)
	calcColor, err := mandelbrot.GetCalculateColor(c)
	if err != nil {
		return err
	}
	dc, err = mandelbrot.GetImage(c, c.CenterX, c.CenterY, c.InitialScale, c.InitialRotation, calcColor)
	if err != nil {
		return err
	}
	initialFrames := int(c.InitialImageTime * float64(c.FramesPerSecond))
	for i := 0; i < initialFrames; i++ {
		png.Encode(os.Stdout, dc.Image())
	}
	zoomFrames := int(c.ZoomTime * float64(c.FramesPerSecond))
	// TODO: check if zoom rotation is right
	zoomRotation := c.Rotation - c.InitialRotation + float64(c.ZoomRotation)*2.0*math.Pi
	scaleProportion := math.Pow(c.Scale/c.InitialScale, 1.0/float64(zoomFrames))
	scale := c.InitialScale
	for i := 0; i < zoomFrames; i++ {
		ratio := float64(i) / float64(zoomFrames)
		scale *= scaleProportion
		rotation := c.InitialRotation + ratio*zoomRotation
		dc, err = mandelbrot.GetImage(c, c.CenterX, c.CenterY, scale, rotation, calcColor)
		if err != nil {
			return err
		}
		png.Encode(os.Stdout, dc.Image())
	}
	dc, err = mandelbrot.GetImage(c, c.CenterX, c.CenterY, c.Scale, c.Rotation, calcColor)
	if err != nil {
		return err
	}
	finalFrames := int(c.FinalImageTime * float64(c.FramesPerSecond))
	for i := 0; i < finalFrames; i++ {
		png.Encode(os.Stdout, dc.Image())
	}
	return nil
}
