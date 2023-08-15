package main

import (
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	windowWidth            = 800
	windowHeight           = 600
	hexRadius              = 25.0
	boardResizeCoefficient = 1.33
	handResizeCoefficient  = 1.0
	beeImagePath           = "assets/bee.png"
)

func drawHexagonAndBee(renderer *sdl.Renderer, x, y int) {
	_hexRadius := hexRadius * boardResizeCoefficient
	centerX := windowWidth/2 + float64(x-y)*_hexRadius*1.5
	centerY := windowHeight/2 + float64(x+y)*_hexRadius*math.Sqrt(3)/2

	// Draw hexagon
	var vx, vy []int16
	for i := 0; i < 6; i++ {
		angle := float64(i) * 2.0 * math.Pi / 6
		vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
		vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
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
	dst := sdl.Rect{X: int32(centerX - _hexRadius), Y: int32(centerY - _hexRadius), W: 2 * int32(_hexRadius), H: 2 * int32(_hexRadius)}
	beeTexture, _ := renderer.CreateTextureFromSurface(imageSurface)
	defer beeTexture.Destroy()
	renderer.Copy(beeTexture, nil, &dst)
}

func DrawHand(renderer *sdl.Renderer, handSize int) {
	_hexRadius := hexRadius * handResizeCoefficient

	// Calculate the total width of the hand
	totalWidth := float64(handSize) * 2 * _hexRadius

	// Calculate the start position (x-coordinate) of the first hexagon
	startX := float64(windowWidth)/2.0 - totalWidth/2.0 + _hexRadius

	// The y-coordinate will be fixed, and will place the hand at the bottom of the screen
	y := windowHeight - int(_hexRadius)

	for i := 0; i < handSize; i++ {
		x := int(startX + float64(i)*(2*_hexRadius))

		// Use the same code as in drawHexagonAndBee to draw each hexagon and bee
		centerX := float64(x)
		centerY := float64(y)

		// Draw hexagon
		var vx, vy []int16
		for j := 0; j < 6; j++ {
			angle := float64(j) * 2.0 * math.Pi / 6
			vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
			vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
		}
		gfx.FilledPolygonColor(renderer, vx, vy, sdl.Color{R: 255, G: 255, B: 255, A: 255})

		// Load the bee image
		imageSurface, _ := img.Load(beeImagePath)
		defer imageSurface.Free()

		// Change all non-transparent pixels to yellow
		var pixelData []byte
		pitch := int(imageSurface.Pitch)
		pixelData = imageSurface.Pixels()
		for h := 0; h < int(imageSurface.H); h++ {
			for w := 0; w < int(imageSurface.W); w++ {
				offset := h*pitch + w*4 // 4 bytes per pixel for ARGB format

				// If pixel is not transparent
				if pixelData[offset+3] > 0 {
					pixelData[offset+0] = 255 // Blue
					pixelData[offset+1] = 255 // Green
					pixelData[offset+2] = 0   // Red
				}
			}
		}

		// Draw bee
		dst := sdl.Rect{X: int32(centerX - _hexRadius), Y: int32(centerY - _hexRadius), W: 2 * int32(_hexRadius), H: 2 * int32(_hexRadius)}
		beeTexture, _ := renderer.CreateTextureFromSurface(imageSurface)
		defer beeTexture.Destroy()
		renderer.Copy(beeTexture, nil, &dst)
	}
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
	drawHexagonAndBee(renderer, 0, 2)

	DrawHand(renderer, 5)

	renderer.Present()
	sdl.Delay(15000) // Wait for 5 seconds to view the result
}
