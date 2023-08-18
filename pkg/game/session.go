package game

type Hand struct {
	Pieces map[PieceType]int
	Color  PieceColor
}

type GameSession struct {
	board    *Board
	white    *Hand
	black    *Hand
	turn     int
	gameOver bool
}

func NewGameSession(handInit func(PieceColor) *Hand) *GameSession {
	gs := &GameSession{
		board:    &Board{},
		white:    handInit(White),
		black:    handInit(Black),
		turn:     0,
		gameOver: false,
	}

	return gs
}

func StandardHand(color PieceColor) *Hand {
	return &Hand{
		Pieces: map[PieceType]int{
			QueenBee:    1,
			Spider:      2,
			Beetle:      2,
			Grasshopper: 3,
			SoldierAnt:  3},
		Color: color,
	}
}

func (gs *GameSession) WhiteToMove() bool {
	return gs.turn%2 == 0
}

func (gs *GameSession) IsGameOver() bool {
	return gs.gameOver
}

func (gs *GameSession) GetBoard() *Board {
	return gs.board
}

func (gs *GameSession) GetWhiteHand() *Hand {
	return gs.white
}

func (gs *GameSession) GetBlackHand() *Hand {
	return gs.black
}

func (gs *GameSession) NextTurn() {
	gs.turn += 1
}
