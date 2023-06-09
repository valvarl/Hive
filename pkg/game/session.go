package game

type Hand map[PieceType]int

type GameSession struct {
	board    *Board
	white    *Hand
	black    *Hand
	turn     int
	gameOver bool
}

func NewGameSession(handInit func() *Hand) *GameSession {
	gs := &GameSession{
		board:    &Board{},
		white:    handInit(),
		black:    handInit(),
		turn:     0,
		gameOver: false,
	}

	return gs
}

func StandardHand() *Hand {
	return &Hand{
		QueenBee:    1,
		Spider:      2,
		Beetle:      2,
		Grasshopper: 2,
		SoldierAnt:  2,
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

func (gs *GameSession) GetBalckHand() *Hand {
	return gs.black
}
