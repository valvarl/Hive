package engine

import (
	"context"
	"hive/pkg/game"
	"math"
	"strconv"
	"sync"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"go.uber.org/zap"
)

const (
	windowWidth            = 800
	windowHeight           = 600
	hexBoardRadius         = 25.0
	hexHandRadius          = 25.0
	imageResizeCoefficient = 1.6
	handFontSize           = 20
	handCircleSize         = 11
	fontPath               = "../assets/NotoSans-Regular.ttf"
)

type UserEngine struct {
	log               *zap.Logger
	init              bool
	active            bool
	title             string
	window            *sdl.Window
	render            *sdl.Renderer
	board             *game.Board
	hand              *game.Hand
	opponentHand      *game.Hand
	color             game.PieceColor
	insectImgPathes   map[game.PieceType]string
	insectImgSurfaces map[game.PieceType]*sdl.Surface
	insectColor       map[game.PieceType]sdl.Color
	pieceTypes        []game.PieceType
	handFont          *ttf.Font
	renderMu          sync.Mutex
	turn              int
	selectedHandPiece int
	selectedPiece     *game.Piece
}

func MakeUserEngine(logger *zap.Logger, title string) *UserEngine {
	return &UserEngine{
		log:   logger,
		init:  false,
		title: title,
		insectImgPathes: map[game.PieceType]string{
			game.QueenBee:    "../assets/bee.png",
			game.Spider:      "../assets/spider.png",
			game.Beetle:      "../assets/beetle.png",
			game.Grasshopper: "../assets/grasshopper.png",
			game.SoldierAnt:  "../assets/ant.png",
		},
		insectColor: map[game.PieceType]sdl.Color{
			game.QueenBee:    {R: 243, G: 218, B: 11, A: 255},
			game.Spider:      {R: 111, G: 79, B: 40, A: 255},
			game.Beetle:      {R: 83, G: 55, B: 122, A: 255},
			game.Grasshopper: {R: 68, G: 148, B: 74, A: 255},
			game.SoldierAnt:  {R: 28, G: 169, B: 201, A: 255},
		},
		insectImgSurfaces: make(map[game.PieceType]*sdl.Surface),
		renderMu:          sync.Mutex{},
		selectedHandPiece: -1,
	}
}

func loadFont(fontPath string, fontSize int) (*ttf.Font, error) {
	font, err := ttf.OpenFont(fontPath, fontSize)
	if err != nil {
		return nil, err
	}

	return font, nil
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

	for pieceType, path := range ue.insectImgPathes {
		// Load the image
		imageSurface, _ := img.Load(path)

		// Change all non-transparent pixels to yellow
		var pixelData []byte
		pitch := int(imageSurface.Pitch)
		pixelData = imageSurface.Pixels()
		for y := 0; y < int(imageSurface.H); y++ {
			for x := 0; x < int(imageSurface.W); x++ {
				offset := y*pitch + x*4 // 4 bytes per pixel for ARGB format

				// If pixel is not transparent
				color := ue.insectColor[pieceType]
				if pixelData[offset+3] > 0 {
					pixelData[offset+0] = color.R
					pixelData[offset+1] = color.G
					pixelData[offset+2] = color.B
				}
			}
		}
		ue.insectImgSurfaces[pieceType] = imageSurface
	}

	ttf.Init()
	ue.handFont, err = loadFont(fontPath, handFontSize)
	if err != nil {
		panic(err)
	}
}

func (ue *UserEngine) Destroy() {
	ue.window.Destroy()
	ue.render.Destroy()

	for _, surface := range ue.insectImgSurfaces {
		surface.Free()
	}

	ue.handFont.Close()
	ttf.Quit()
}

func (ue *UserEngine) DrawBoard(mouseX, mouseY int, isClicking bool, shiftX, shiftY float64) (selectedPosition *game.Position, selectEmptyRoom bool) {
	_hexRadius := hexBoardRadius * imageResizeCoefficient

	for _, piece := range ue.board.Pieces {
		centerX := windowWidth/2 + float64(piece.Position.X-piece.Position.Y)*_hexRadius*1.5 + shiftX
		centerY := windowHeight/2 + float64(piece.Position.X+piece.Position.Y)*_hexRadius*math.Sqrt(3)/2 + shiftY

		// Draw hexagon
		var vx, vy []int16
		for i := 0; i < 6; i++ {
			angle := float64(i) * 2.0 * math.Pi / 6
			vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
			vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
		}
		var pieceColor sdl.Color
		if piece.Color == game.White {
			pieceColor = sdl.Color{R: 255, G: 255, B: 240, A: 255}
		} else {
			pieceColor = sdl.Color{R: 24, G: 23, B: 28, A: 255}
		}
		gfx.FilledPolygonColor(ue.render, vx, vy, pieceColor)

		placeQueen := ue.turn >= 6 && ue.hand.Pieces[game.QueenBee] != 0

		if !placeQueen && mouseX != -1 && mouseY != -1 && piece.Color == ue.color {
			if ue.pointInsidePolygon(int16(mouseX), int16(mouseY), vx, vy) {
				gfx.FilledPolygonColor(ue.render, vx, vy, sdl.Color{R: 60, G: 60, B: 60, A: 20})
				if isClicking {
					selectedPosition = &game.Position{X: piece.Position.X, Y: piece.Position.Y}
					selectEmptyRoom = false
				}
			}
		}

		// Draw insect
		dst := sdl.Rect{X: int32(centerX - hexBoardRadius), Y: int32(centerY - hexBoardRadius), W: 2 * int32(hexBoardRadius), H: 2 * int32(hexBoardRadius)}
		beeTexture, _ := ue.render.CreateTextureFromSurface(ue.insectImgSurfaces[piece.Type])
		defer beeTexture.Destroy()
		ue.render.Copy(beeTexture, nil, &dst)
	}

	var possibleMoves []game.Position
	var color sdl.Color
	if ue.selectedHandPiece != -1 {
		possibleMoves = game.AvailableToPlace(ue.board, ue.color)
		color = ue.insectColor[ue.pieceTypes[ue.selectedHandPiece]]
	} else if ue.selectedPiece != nil {
		possibleMoves = game.AvailableToMove(ue.board, ue.selectedPiece)
		color = ue.insectColor[ue.selectedPiece.Type]
	} else {
		return selectedPosition, selectEmptyRoom
	}

	for _, position := range possibleMoves {
		centerX := windowWidth/2 + float64(position.X-position.Y)*_hexRadius*1.5 + shiftX
		centerY := windowHeight/2 + float64(position.X+position.Y)*_hexRadius*math.Sqrt(3)/2 + shiftY

		var vx, vy []int16
		for j := 0; j < 6; j++ {
			angle := float64(j) * 2.0 * math.Pi / 6
			vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
			vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
		}

		color.A = 128
		if ue.pointInsidePolygon(int16(mouseX), int16(mouseY), vx, vy) {
			color.A = 255
			selectedPosition = &game.Position{X: position.X, Y: position.Y}
			selectEmptyRoom = true
		}
		for j := 0; j < 5; j++ {
			gfx.ThickLineColor(ue.render, int32(vx[j]), int32(vy[j]), int32(vx[j+1]), int32(vy[j+1]), 2, color)
		}
		gfx.ThickLineColor(ue.render, int32(vx[0]), int32(vy[0]), int32(vx[5]), int32(vy[5]), 2, color)
	}

	return selectedPosition, selectEmptyRoom
}

func (ue *UserEngine) drawText(text string, x, y int, color sdl.Color) {
	surface, err := ue.handFont.RenderUTF8Solid(text, color)
	if err != nil {
		panic(err)
	}
	defer surface.Free()
	texture, err := ue.render.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()
	ue.render.Copy(texture, nil, &sdl.Rect{X: int32(x), Y: int32(y), W: surface.W, H: surface.H})
}

func (ue *UserEngine) DrawHand(mouseX, mouseY int, isClicking bool) (selectedPiece int) {
	selectedPiece = -1
	_hexRadius := hexHandRadius * imageResizeCoefficient

	// Calculate the total width of the hand
	totalWidth := float64(len(ue.hand.Pieces)) * 2 * _hexRadius

	// Calculate the start position (x-coordinate) of the first hexagon
	startX := float64(windowWidth)/2.0 - totalWidth/2.0 + _hexRadius

	// The y-coordinate will be fixed, and will place the hand at the bottom of the screen
	y := windowHeight - int(_hexRadius)

	for i, pieceType := range ue.pieceTypes {
		x := int(startX + float64(i)*(2*_hexRadius))

		centerX := float64(x)
		centerY := float64(y)

		// Draw hexagon
		var vx, vy []int16
		for j := 0; j < 6; j++ {
			angle := float64(j) * 2.0 * math.Pi / 6
			vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
			vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
		}

		var pieceColor sdl.Color
		if ue.color == game.White {
			pieceColor = sdl.Color{R: 255, G: 255, B: 240, A: 255}
		} else {
			pieceColor = sdl.Color{R: 24, G: 23, B: 28, A: 255}
		}
		gfx.FilledPolygonColor(ue.render, vx, vy, pieceColor)

		placeQueen := ue.turn >= 6 && ue.hand.Pieces[game.QueenBee] != 0

		if placeQueen && ue.pieceTypes[i] == game.QueenBee || !placeQueen && ue.hand.Pieces[ue.pieceTypes[i]] > 0 {
			var color sdl.Color = ue.insectColor[ue.pieceTypes[i]]
			color.A = 0
			if mouseX != -1 && mouseY != -1 {
				if ue.pointInsidePolygon(int16(mouseX), int16(mouseY), vx, vy) {
					if isClicking {
						color.A = 255
						ue.selectedPiece = nil
					} else {
						color.A = 128
					}
					selectedPiece = i
				}
			}

			if ue.selectedHandPiece == i {
				color.A = 255
			}

			for j := 0; j < 5; j++ {
				gfx.ThickLineColor(ue.render, int32(vx[j]), int32(vy[j]), int32(vx[j+1]), int32(vy[j+1]), 2, color)
			}
			gfx.ThickLineColor(ue.render, int32(vx[0]), int32(vy[0]), int32(vx[5]), int32(vy[5]), 2, color)
		}

		// Draw insect
		dst := sdl.Rect{X: int32(centerX - hexHandRadius), Y: int32(centerY - hexHandRadius), W: 2 * int32(hexHandRadius), H: 2 * int32(hexHandRadius)}
		texture, _ := ue.render.CreateTextureFromSurface(ue.insectImgSurfaces[pieceType])
		defer texture.Destroy()
		ue.render.Copy(texture, nil, &dst)

		// Draw circle
		circleX := x + int(_hexRadius) - int(handCircleSize)
		circleY := y + int(_hexRadius*math.Sqrt(3)/2) - int(handCircleSize)
		gfx.FilledCircleColor(ue.render, int32(circleX), int32(circleY), int32(handCircleSize), sdl.Color{R: 220, G: 220, B: 220, A: 200})

		textColor := sdl.Color{R: 60, G: 60, B: 60, A: 255}
		numberText := strconv.Itoa((ue.hand.Pieces)[ue.pieceTypes[i]])
		w, h, err := ue.handFont.SizeUTF8(numberText)
		if err != nil {
			panic(err)
		}
		ue.drawText(numberText, circleX-w/2, circleY-h/2, textColor)
	}
	return selectedPiece
}

func (ue *UserEngine) DrawOpponentHand() {
	_hexRadius := hexHandRadius * imageResizeCoefficient

	// Calculate the total width of the hand
	totalWidth := float64(len(ue.hand.Pieces)) * 2 * _hexRadius

	// Calculate the start position (x-coordinate) of the first hexagon
	startX := float64(windowWidth)/2.0 - totalWidth/2.0 + _hexRadius

	// The y-coordinate will be fixed, and will place the hand at the bottom of the screen
	y := int(_hexRadius)

	for i, pieceType := range ue.pieceTypes {
		x := int(startX + float64(i)*(2*_hexRadius))

		centerX := float64(x)
		centerY := float64(y)

		// Draw hexagon
		var vx, vy []int16
		for j := 0; j < 6; j++ {
			angle := float64(j) * 2.0 * math.Pi / 6
			vx = append(vx, int16(centerX+_hexRadius*math.Cos(angle)))
			vy = append(vy, int16(centerY+_hexRadius*math.Sin(angle)))
		}

		var pieceColor sdl.Color
		if ue.color == game.White {
			pieceColor = sdl.Color{R: 24, G: 23, B: 28, A: 255}
		} else {
			pieceColor = sdl.Color{R: 255, G: 255, B: 240, A: 255}
		}
		gfx.FilledPolygonColor(ue.render, vx, vy, pieceColor)

		// Draw insect
		dst := sdl.Rect{X: int32(centerX - hexHandRadius), Y: int32(centerY - hexHandRadius), W: 2 * int32(hexHandRadius), H: 2 * int32(hexHandRadius)}
		texture, _ := ue.render.CreateTextureFromSurface(ue.insectImgSurfaces[pieceType])
		defer texture.Destroy()
		ue.render.Copy(texture, nil, &dst)

		// Draw circle
		circleX := x + int(_hexRadius) - int(handCircleSize)
		circleY := y + int(_hexRadius*math.Sqrt(3)/2) - int(handCircleSize)
		gfx.FilledCircleColor(ue.render, int32(circleX), int32(circleY), int32(handCircleSize), sdl.Color{R: 220, G: 220, B: 220, A: 200})

		textColor := sdl.Color{R: 60, G: 60, B: 60, A: 255}
		numberText := strconv.Itoa((ue.opponentHand.Pieces)[ue.pieceTypes[i]])
		w, h, err := ue.handFont.SizeUTF8(numberText)
		if err != nil {
			panic(err)
		}
		ue.drawText(numberText, circleX-w/2, circleY-h/2, textColor)
	}
}

func (ue *UserEngine) Start(ctx context.Context, board *game.Board, hand, opponentHand *game.Hand, engineResponse chan *game.Move) {
	ue.renderMu.Lock()
	ue.board = board
	ue.hand = hand
	ue.opponentHand = opponentHand
	ue.color = ue.hand.Color
	ue.active = true
	ue.renderMu.Unlock()

	if !ue.init {
		ue.Init()
	}

	for _, pt := range []game.PieceType{game.QueenBee, game.Spider, game.Beetle, game.Grasshopper, game.SoldierAnt} {
		if _, ok := (ue.hand.Pieces)[pt]; ok {
			ue.pieceTypes = append(ue.pieceTypes, pt)
		}
	}

	// Объявите переменные для отслеживания состояния клика и перетаскивания
	var isClicking bool
	var isDragging bool
	var startX, startY int
	var hoverX, hoverY int
	draggingDeactivate := false
	startShiftX := 0.0
	startShiftY := 0.0
	shiftX := 0.0
	shiftY := 0.0
	threshold := 1.0

	// Основной цикл событий
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Обработка событий
			ue.renderMu.Lock()
			if ue.active {
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
									print("endDragging")
								} else {
									// Обработка обычного клика
									print(ue.window.GetTitle(), ": ", t.X, " ", t.Y, "\n")
								}
							}
							draggingDeactivate = false
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
						}
					}
				}
			}

			// Очистка экрана и отрисовка объектов
			ue.render.SetDrawColor(128, 128, 128, 255)
			ue.render.Clear()

			ue.DrawOpponentHand()

			if ue.active {
				if !isClicking {
					ue.DrawHand(hoverX, hoverY, isClicking)
				} else {
					if selectedPiece := ue.DrawHand(startX, startY, isClicking); selectedPiece != -1 {
						ue.selectedHandPiece = selectedPiece
						ue.selectedPiece = nil
						draggingDeactivate = true
					}
				}

				if !isClicking {
					ue.DrawBoard(hoverX, hoverY, isClicking, shiftX, shiftY)
				} else {
					position, selectEmptyRoom := ue.DrawBoard(startX, startY, isClicking, shiftX, shiftY)

					if position != nil {
						draggingDeactivate = true
						if selectEmptyRoom {
							var movePlayed *game.Move
							if ue.selectedHandPiece != -1 {
								piecePlaced := &game.Piece{Position: *position, Type: ue.pieceTypes[ue.selectedHandPiece], Color: ue.color, Placed: false}
								ue.board.Pieces = append(ue.board.Pieces, piecePlaced)
								ue.hand.Pieces[ue.pieceTypes[ue.selectedHandPiece]] -= 1
								movePlayed = &game.Move{Piece: piecePlaced, Position: position}
							} else {
								piece := &game.Piece{Position: ue.selectedPiece.Position, Type: ue.selectedPiece.Type, Color: ue.selectedPiece.Color, Placed: true}
								movePlayed = &game.Move{Piece: piece, Position: position}
								(*ue.selectedPiece).Position = *position
							}
							ue.selectedHandPiece = -1
							ue.selectedPiece = nil
							ue.active = false
							draggingDeactivate = false
							engineResponse <- movePlayed
						} else {
							ue.log.Info("Position selected", zap.Any("position", position), zap.Any("deck", ue.board.Pieces))

							for _, p := range ue.board.Pieces {
								if p.Position.X == position.X && p.Position.Y == position.Y {
									ue.selectedPiece = p
									break
								}
							}
							ue.selectedHandPiece = -1
						}
					}
				}
				// Отображение результата на экране
				ue.render.Present()

				// Задержка
				sdl.Delay(16) // Примерно 60 кадров в секунду

			} else {
				ue.DrawHand(hoverX, hoverY, isClicking)
				ue.DrawBoard(hoverX, hoverY, isClicking, shiftX, shiftY)

				// Отображение результата на экране
				ue.render.Present()

				// Задержка
				sdl.Delay(32)
			}
			ue.renderMu.Unlock()
		}
	}
}

func (ue *UserEngine) Update(board *game.Board, hand, opponentHand *game.Hand, turn int) {
	ue.renderMu.Lock()
	ue.board = board
	ue.hand = hand
	ue.opponentHand = opponentHand
	ue.turn = turn
	ue.selectedHandPiece = -1
	ue.selectedPiece = nil
	ue.active = true
	ue.window.Raise()
	ue.renderMu.Unlock()
}

func (ue *UserEngine) pointInsidePolygon(x, y int16, verticesX, verticesY []int16) bool {
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
