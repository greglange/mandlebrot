package mandelbrot

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
)

type Config struct {
	RunType          string  `yaml:"run_type" env:"run_type" env-default:"image"`
	MaxIteration     int     `yaml:"max_iteration" env:"max_iteration" env-default:"1000"`
	ColorsFilePath   string  `yaml:"colors_file_path" env:"colors_file_path" env-default:"default.pal"`
	ColorCalc        string  `yaml:"color_calc" env:"color_calc" env-default:"smooth"`
	ImageWidth       int     `yaml:"image_width" env:"image_width" env-default:"1920"`
	ImageHeight      int     `yaml:"image_height" env:"image_height" env-default:"1080"`
	CenterX          float64 `yaml:"center_x" env:"center_x" env-default:"0.0"`
	CenterY          float64 `yaml:"center_y" env:"center_y" env-default:"0.0"`
	Scale            float64 `yaml:"scale" env:"scale" env-default:"0.003"`
	Rotation         float64 `yaml:"rotation" env:"rotation" env-default:"0.0"`
	XLines           string  `yaml:"x_lines" env:"x_lines" env-default:""`
	YLines           string  `yaml:"y_lines" env:"y_lines" env-default:""`
	FramesPerSecond  int     `yaml:"frames_per_second" env:"frames_per_second" env-default:"60"`
	InitialImageTime float64 `yaml:"initial_image_time" env:"initial_image_time" env-default:"1.0"`
	InitialScale     float64 `yaml:"initial_scale" env:"initial_scale" env-default:"0.003"`
	InitialRotation  float64 `yaml:"initial_rotation" env:"initial_rotation" env-default:"0.0"`
	ZoomTime         float64 `yaml:"zoom_time" env:"zoom_time" env-default:"10.0"`
	ZoomRotation     int     `yaml:"zoom_rotation" env:"zoom_rotation" env-default:"0"`
	FinalImageTime   float64 `yaml:"final_image_time" env:"final_image_time" env-default:"1.0"`
}

// TODO: add doc
func GetCalculateColor(config Config) (func(bool, int, float64) Color, error) {
	colors, err := LoadColors(config)
	if err != nil {
		return nil, err
	}
	if config.ColorCalc == "smooth" {
		// https://en.wikipedia.org/wiki/Plotting_algorithms_for_the_Mandelbrot_set#Continuous_(smooth)_coloring
		return func(inSet bool, escapeIteration int, escapeValue float64) Color {
			if inSet {
				return Color{0, 0, 0}
			}
			i := int(math.Floor(escapeValue))
			s := math.Mod(escapeValue, 1.0)
			c1 := colors[(i+1)%len(colors)]
			c2 := colors[i%len(colors)]
			r := int(float64(c1.R)*s + float64(c2.R)*(1.0-s))
			g := int(float64(c1.G)*s + float64(c2.G)*(1.0-s))
			b := int(float64(c1.B)*s + float64(c2.B)*(1.0-s))
			return Color{r, g, b}
		}, nil
	}
	return nil, errors.New("Invalid name")
}

// TODO: add doc
func CalculateSet(config Config, a, b float64) (bool, int, float64) {
	var iteration int
	var aCurrent, bCurrent float64
	for iteration < config.MaxIteration {
		aSquared := aCurrent * aCurrent
		bSquared := bCurrent * bCurrent
		current := aSquared + bSquared
		aTemp := aSquared - bSquared + a
		bCurrent = 2*aCurrent*bCurrent + b
		aCurrent = aTemp
		if current > 64.0 {
			break
		}
		iteration += 1
	}
	var color float64
	if iteration < config.MaxIteration {
		logZn := math.Log(aCurrent*aCurrent+bCurrent*bCurrent) / 2
		nu := math.Log(logZn/math.Log(2)) / math.Log(2)
		color = float64(iteration) + 1.0 - nu
		if color < 0.0 {
			color = 0.0
		}
	}
	return iteration == config.MaxIteration, iteration, color
}

type Color struct {
	R, G, B int
}

func (c *Color) valid() bool {
	if c.R < 0 || c.R > 255 {
		return false
	}
	if c.G < 0 || c.G > 255 {
		return false
	}
	if c.B < 0 || c.B > 255 {
		return false
	}
	return true
}

func LoadColors(config Config) ([]Color, error) {
	var err error
	file, err := os.Open(config.ColorsFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	colors := make([]Color, 0)
	lineNumber := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber += 1
		if strings.HasPrefix(line, "#") {
			continue
		}
		color := Color{}
		n, err := fmt.Sscanf(line, "%d %d %d", &color.R, &color.G, &color.B)
		if err != nil {
			return nil, errors.New(err.Error() + " on line " + strconv.Itoa(lineNumber))
		}
		if n != 3 {
			return nil, errors.New("Unexpected number of colors on line " + strconv.Itoa(lineNumber))
		}
		if !color.valid() {
			return nil, errors.New("Invalid color on line " + strconv.Itoa(lineNumber))
		}
		colors = append(colors, color)
	}
	return colors, nil
}

func ImageToPlane(config Config, x, y int) (float64, float64) {
	a := (float64(x) - float64(config.ImageWidth-1)/2) * config.Scale
	b := (float64(config.ImageHeight-1)/2 - float64(y)) * config.Scale
	rotA := math.Cos(config.Rotation)*a - math.Sin(config.Rotation)*b
	rotB := math.Sin(config.Rotation)*a + math.Cos(config.Rotation)*b
	return config.CenterX + rotA, config.CenterY + rotB
}

func PlaneToImage(config Config, x, y float64) (int, int) {
	rotA := x - config.CenterX
	rotB := y - config.CenterY
	a := math.Cos(config.Rotation)*rotA + math.Sin(config.Rotation)*rotB
	b := -math.Sin(config.Rotation)*rotA + math.Cos(config.Rotation)*rotB
	aa := a/config.Scale + float64(config.ImageWidth-1)/2
	bb := -b/config.Scale + float64(config.ImageHeight-1)/2
	return int(math.Round(aa)), int(math.Round(bb))
}

func floatsFromString(source string) ([]float64, error) {
	floats := []float64{}
	source = strings.TrimSpace(source)
	if len(source) == 0 {
		return floats, nil
	}
	for _, v := range strings.Split(source, ",") {
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, err
		}
		floats = append(floats, f)
	}
	return floats, nil
}

func drawLine(config Config, dc *gg.Context, x1, y1, x2, y2 float64) {
	x1_, y1_ := PlaneToImage(config, x1, y1)
	x2_, y2_ := PlaneToImage(config, x2, y2)
	dc.SetRGBA255(255, 255, 255, 255)
	dc.DrawLine(float64(x1_), float64(y1_), float64(x2_), float64(y2_))
	dc.Stroke()
}

func DrawLines(config Config, dc *gg.Context) error {
	xLines, err := floatsFromString(config.XLines)
	if err != nil {
		return err
	}
	yLines, err := floatsFromString(config.YLines)
	if err != nil {
		return err
	}
	if len(xLines) == 0 && len(yLines) == 0 {
		return nil
	}
	type corner struct{ x, y int }
	corners := []corner{
		corner{0, 0},
		corner{0, config.ImageHeight - 1},
		corner{config.ImageWidth - 1, 0},
		corner{config.ImageWidth - 1, config.ImageHeight - 1},
	}
	xs := []float64{}
	ys := []float64{}
	for _, c := range corners {
		x, y := ImageToPlane(config, c.x, c.y)
		xs = append(xs, x)
		ys = append(ys, y)
	}
	for _, x := range xLines {
		if slices.Min(xs) <= x && x <= slices.Max(xs) {
			drawLine(config, dc, x, slices.Min(ys), x, slices.Max(ys))
		}
	}
	for _, y := range yLines {
		if slices.Min(ys) <= y && y <= slices.Max(ys) {
			drawLine(config, dc, slices.Min(xs), y, slices.Max(xs), y)
		}
	}
	return nil
}

func GetImage(config Config, centerX, centerY, scale, rotation float64, calcColor func(bool, int, float64) Color) (*gg.Context, error) {
	dc := gg.NewContext(config.ImageWidth, config.ImageHeight)
	numCPU := runtime.NumCPU()
	workerCount := numCPU - 2
	if workerCount < 1 {
		workerCount = 1
	}
	type job struct {
		x int
		y int
		a float64
		b float64
	}
	type result struct {
		x     int
		y     int
		color Color
	}
	calc := func(x, y int, a, b float64) result {
		rotA := math.Cos(rotation)*a - math.Sin(rotation)*b
		rotB := math.Sin(rotation)*a + math.Cos(rotation)*b
		tranA := centerX + rotA
		tranB := centerY + rotB
		inSet, escapeIteration, escapeValue := CalculateSet(config, tranA, tranB)
		color := calcColor(inSet, escapeIteration, escapeValue)
		return result{x, y, color}
	}
	worker := func(jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
		defer wg.Done()
		for j := range jobs {
			results <- calc(j.x, j.y, j.a, j.b)
		}
	}
	collectResults := func(results <-chan result, wg *sync.WaitGroup) {
		defer wg.Done()
		for r := range results {
			dc.SetRGBA255(r.color.R, r.color.G, r.color.B, 255)
			dc.SetPixel(r.x, r.y)
		}
	}
	jobs := make(chan job, workerCount*100)
	results := make(chan result, workerCount*100)
	var jobsWaitGroup sync.WaitGroup
	jobsWaitGroup.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker(jobs, results, &jobsWaitGroup)
	}
	var resultsWaitGroup sync.WaitGroup
	resultsWaitGroup.Add(1)
	go collectResults(results, &resultsWaitGroup)
	for x := 0; x < config.ImageWidth; x++ {
		a := (float64(x) - float64(config.ImageWidth-1)/2) * scale
		for y := 0; y < config.ImageHeight; y++ {
			b := (float64(config.ImageHeight-1)/2 - float64(y)) * scale
			jobs <- job{x, y, a, b}
		}
	}
	close(jobs)
	jobsWaitGroup.Wait()
	close(results)
	resultsWaitGroup.Wait()
	return dc, nil
}
