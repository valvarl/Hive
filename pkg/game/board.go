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
	Placed   bool
}

type Board struct {
	Pieces []Piece
}

type Move struct {
	Piece    *Piece
	Position *Position
}

func IsPositionNeignbour(lhs, rhs Position) bool {
	relative_position := Position{X: lhs.X - rhs.X, Y: lhs.Y - rhs.Y}
	if relative_position.X == 1 {
		if relative_position.Y == 0 || relative_position.Y == 1 {
			return true
		}
	} else if relative_position.X == 0 {
		if relative_position.Y == 1 || relative_position.Y == -1 {
			return true
		}
	} else if relative_position.X == -1 {
		if relative_position.Y == 0 || relative_position.Y == -1 {
			return true
		}
	}
	return false
}

func AvailableToPlace(board *Board, color PieceColor) []Position {
	var positions []Position
	if len(board.Pieces) == 0 {
		positions = append(positions, Position{0, 0})
	} else if len(board.Pieces) == 1 {
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				if (i != 0 || j != 0) && i+j != 0 {
					positions = append(positions, Position{i, j})
				}
			}
		}
	} else {
		for _, piece := range board.Pieces {
			if piece.Color == color {
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						if (i != 0 || j != 0) && i+j != 0 {
							checkPosition := Position{piece.Position.X + i, piece.Position.Y + j}
							positionAvailable := true

							for _, k := range board.Pieces {
								if checkPosition.X == k.Position.X && checkPosition.Y == k.Position.Y {
									positionAvailable = false
									break
								}
								if k.Color != color {
									if IsPositionNeignbour(checkPosition, k.Position) {
										positionAvailable = false
										break
									}
								}
							}

							if positionAvailable {
								positions = append(positions, checkPosition)
							}
						}
					}
				}
			}
		}
	}
	return positions
}
