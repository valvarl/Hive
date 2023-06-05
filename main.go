package main

import (
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth  = 800
	windowHeight = 600
	hexRadius    = 50.0
	beeImagePath = "assets/bee.png"
)

func drawHexagonAndBee(renderer *sdl.Renderer, x, y int) {
	centerX := windowWidth/2 + float64(x-y)*hexRadius*1.5
	centerY := windowHeight/2 + float64(x+y)*hexRadius*math.Sqrt(3)/2

	// Draw hexagon
	var vx, vy []int16
	for i := 0; i < 6; i++ {
		angle := float64(i) * 2.0 * math.Pi / 6
		vx = append(vx, int16(centerX+hexRadius*math.Cos(angle)))
		vy = append(vy, int16(centerY+hexRadius*math.Sin(angle)))
	}
	gfx.FilledPolygonColor(renderer, vx, vy, sdl.Color{R: 255, G: 255, B: 255, A: 255})

	// Load the bee image
	imageSurface, _ := img.Load(beeImagePath)
	defer imageSurface.Free()

	// Change all non-transparent pixels to yellow
	var pixelData []byte
	pitch := int(imageSurface.Pitch)
	pixelData = imageSurface.Pixels()
	for y := 0; y < int(imageSurface.H); y++ {
		for x := 0; x < int(imageSurface.W); x++ {
			offset := y*pitch + x*4 // 4 bytes per pixel for ARGB format

			// If pixel is not transparent
			if pixelData[offset+3] > 0 {
				pixelData[offset+0] = 255 // Blue
				pixelData[offset+1] = 255 // Green
				pixelData[offset+2] = 0   // Red
			}
		}
	}

	// Draw bee
	dst := sdl.Rect{X: int32(centerX - hexRadius), Y: int32(centerY - hexRadius), W: 2 * int32(hexRadius), H: 2 * int32(hexRadius)}
	beeTexture, _ := renderer.CreateTextureFromSurface(imageSurface)
	defer beeTexture.Destroy()
	renderer.Copy(beeTexture, nil, &dst)
}

func main() {
	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	window, _ := sdl.CreateWindow("Hexagon", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	defer window.Destroy()

	renderer, _ := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	defer renderer.Destroy()

	renderer.SetDrawColor(128, 128, 128, 255) // Set to gray color
	renderer.Clear()                          // Fill the entire screen with gray color

	drawHexagonAndBee(renderer, 0, 0)
	drawHexagonAndBee(renderer, 1, 1)

	renderer.Present()
	sdl.Delay(5000) // Wait for 5 seconds to view the result
}
