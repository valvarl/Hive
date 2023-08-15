package engine

import (
	"context"
	"hive/pkg/game"
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"go.uber.org/zap"
)

const (
	windowWidth  = 800
	windowHeight = 600
	hexRadius    = 50.0
	beeImagePath = "assets/bee.png"
)

type UserEngine struct {
	log          *zap.Logger
	init         bool
	title        string
	window       *sdl.Window
	render       *sdl.Renderer
	board        *game.Board
	hand         *game.Hand
	opponentHand *game.Hand
}

func MakeUserEngine(logger *zap.Logger, title string) *UserEngine {
	return &UserEngine{log: logger, init: false, title: title}
}

func (ue *UserEngine) Init() {
	window, err := sdl.CreateWindow(ue.title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(windowWidth), int32(windowHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	ue.window = window
	ue.render = renderer
}

func (ue *UserEngine) Destroy() {
	ue.window.Destroy()
	ue.render.Destroy()
}

func (ue *UserEngine) MakeMove(ctx context.Context, board *game.Board, hand, opponentHand *game.Hand) *game.Move {
	if !ue.init {
		ue.Init()
	}
	ue.board = board
	ue.hand = hand
	ue.opponentHand = opponentHand
	err := ue.UpdateRender()
	if err != nil {
		ue.log.Error("Ошибка рендера", zap.Error(err))
		return nil
	}
	// for {
	// 	select {
	// 	case <- ctx.Done():
	// 		case <-
	// 	}
	// }
	return nil
}

func (ue *UserEngine) HandleEvent() {

}

func (ue *UserEngine) drawHexagon(x, y int) {
	centerX := windowWidth/2 + float64(x-y)*hexRadius*1.5
	centerY := windowHeight/2 + float64(x+y)*hexRadius*math.Sqrt(3)/2

	// Draw hexagon
	var vx, vy []int16
	for i := 0; i < 6; i++ {
		angle := float64(i) * 2.0 * math.Pi / 6
		vx = append(vx, int16(centerX+hexRadius*math.Cos(angle)))
		vy = append(vy, int16(centerY+hexRadius*math.Sin(angle)))
	}
	gfx.FilledPolygonColor(ue.render, vx, vy, sdl.Color{R: 255, G: 255, B: 255, A: 255})

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
	beeTexture, _ := ue.render.CreateTextureFromSurface(imageSurface)
	defer beeTexture.Destroy()
	ue.render.Copy(beeTexture, nil, &dst)
}

func (ue *UserEngine) UpdateRender() error {
	return nil
}
