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

var (
	startShiftX       = 0.0
	startShiftY       = 0.0
	shiftX            = 0.0
	shiftY            = 0.0
	selectedHandPiece = -1
)

func drawHexagonAndBee(renderer *sdl.Renderer, x, y int) {
	_hexRadius := hexRadius * boardResizeCoefficient
	centerX := windowWidth/2 + float64(x-y)*_hexRadius*1.5 + shiftX
	centerY := windowHeight/2 + float64(x+y)*_hexRadius*math.Sqrt(3)/2 + shiftY

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

func DrawHand(renderer *sdl.Renderer, handSize int, mouseX, mouseY int, isClicking bool) (selectedPiece int) {
	selectedPiece = -1
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

		var color sdl.Color
		if mouseX != -1 && mouseY != -1 {
			if pointInsidePolygon(int16(mouseX), int16(mouseY), vx, vy) {
				if isClicking {
					color = sdl.Color{R: 0, G: 255, B: 0, A: 255}
				} else {
					color = sdl.Color{R: 60, G: 60, B: 60, A: 255}
				}
				selectedPiece = i
			}
		}

		if selectedHandPiece == i {
			color = sdl.Color{R: 0, G: 255, B: 0, A: 255}
		}

		gfx.PolygonColor(renderer, vx, vy, color)

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
	return selectedPiece
}

func DrawBoard() {

}

func pointInsidePolygon(x, y int16, verticesX, verticesY []int16) bool {
	numVertices := len(verticesX)
	if numVertices != len(verticesY) || numVertices < 3 {
		return false
	}

	intersections := 0
	for i := 0; i < numVertices; i++ {
		x1, y1 := verticesX[i], verticesY[i]
		x2, y2 := verticesX[(i+1)%numVertices], verticesY[(i+1)%numVertices]

		if ((y1 > y) != (y2 > y)) && (x < (x2-x1)*(y-y1)/(y2-y1)+x1) {
			intersections++
		}
	}

	return intersections%2 == 1
}

func main() {
	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	window, _ := sdl.CreateWindow("Hexagon", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	defer window.Destroy()

	renderer, _ := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	defer renderer.Destroy()

	// Объявите переменные для отслеживания состояния клика и перетаскивания
	var isClicking bool
	var isDragging bool
	var startX, startY int
	var hoverX, hoverY int
	draggingDeactivate := false
	threshold := 3.0

	// Основной цикл событий
	for {
		// Обработка событий
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				// Обработка выхода из приложения (по крестику)
				return
			case *sdl.MouseButtonEvent:
				if t.Button == sdl.BUTTON_LEFT {
					if t.State == sdl.PRESSED {
						// Обработка начала клика
						isClicking = true
						startX, startY = int(t.X), int(t.Y)
					} else if t.State == sdl.RELEASED {
						// Обработка окончания клика
						isClicking = false

						if isDragging {
							// Завершение перетаскивания
							isDragging = false
							draggingDeactivate = false
							print("endDragging")
						} else {
							// Обработка обычного клика
							draggingDeactivate = false
							print(t.X, " ", t.Y, "\n")
						}
					}
				}
			case *sdl.MouseMotionEvent:
				if isClicking {
					// Обработка движения мыши во время клика (перетаскивание)
					if !isDragging {
						// Проверьте, началось ли перетаскивание (например, с определенным порогом смещения)
						deltaX := int(t.X) - startX
						deltaY := int(t.Y) - startY
						if math.Abs(float64(deltaX)) > threshold || math.Abs(float64(deltaY)) > threshold {
							isDragging = true
							// Дополнительные действия при начале перетаскивания
							print("isDragging")
							startShiftX = shiftX
							startShiftY = shiftY
						}
					}

					if isDragging && !draggingDeactivate {
						// Дополнительные действия во время перетаскивания
						shiftX = float64(int(t.X) + int(startShiftX) - startX)
						shiftY = float64(int(t.Y) + int(startShiftY) - startY)
					}
				} else {
					hoverX = int(t.X)
					hoverY = int(t.Y)
					// print(t.X, " ", t.Y, "\n")
				}
			}
		}

		// Очистка экрана и отрисовка объектов
		renderer.SetDrawColor(128, 128, 128, 255)
		renderer.Clear()

		drawHexagonAndBee(renderer, 0, 0)
		drawHexagonAndBee(renderer, 0, 2)

		if !isClicking {
			DrawHand(renderer, 5, hoverX, hoverY, isClicking)
		} else {
			if selectedPiece := DrawHand(renderer, 5, startX, startY, isClicking); selectedPiece != -1 {
				selectedHandPiece = selectedPiece
				draggingDeactivate = true
			}
		}

		// Отображение результата на экране
		renderer.Present()

		// Задержка
		sdl.Delay(16) // Примерно 60 кадров в секунду
	}
}
