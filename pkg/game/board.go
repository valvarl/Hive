package game

type Position struct {
	X int
	Y int
}

type PieceType int

const (
	QueenBee PieceType = iota
	Spider
	Beetle
	Grasshopper
	SoldierAnt
	// дополнительные типы для расширений
)

type PieceColor int

const (
	White PieceColor = iota
	Black
)

type Piece struct {
	Position Position
	Type     PieceType
	Color    PieceColor
}

type Board struct {
	Pieces []Piece
}
