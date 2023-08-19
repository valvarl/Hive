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
	Level    int
}

type Board struct {
	Pieces []*Piece
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

type Data struct {
	Level int
	Color PieceColor
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
		set := map[Position]Data{}
		for _, piece := range board.Pieces {
			if data, ok := set[piece.Position]; ok {
				if piece.Level > data.Level {
					set[piece.Position] = Data{piece.Level, piece.Color}
				}
			} else {
				set[piece.Position] = Data{piece.Level, piece.Color}
			}
		}

		for pos := range set {
			if set[pos].Color == color {
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						if (i != 0 || j != 0) && i+j != 0 {
							checkPosition := Position{pos.X + i, pos.Y + j}
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

func CanSqueezeThrough(board *Board, lhs, rhs Position) bool {
	commonNeighbours := 0
	for _, piece := range board.Pieces {
		if IsPositionNeignbour(lhs, piece.Position) && IsPositionNeignbour(rhs, piece.Position) {
			commonNeighbours += 1
		}
	}
	return commonNeighbours != 2
}

func CanMove(board *Board, piece *Piece) bool {
	visited := map[Position]bool{}
	for _, p := range board.Pieces {
		if p.Position.X != piece.Position.X || p.Position.Y != piece.Position.Y {
			visited[p.Position] = false
		} else if p.Level > piece.Level {
			return false
		}
	}

	if len(visited) > 0 {
		l1 := &[]Position{}
		l2 := &[]Position{}
		for key := range visited {
			*l1 = append(*l1, key)
			visited[key] = true
			break
		}

		for len(*l1) > 0 {
			for _, p := range *l1 {
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						if (i != 0 || j != 0) && i+j != 0 {
							pos := Position{X: p.X + i, Y: p.Y + j}
							vis, ok := visited[pos]
							if ok && !vis {
								visited[pos] = true
								*l2 = append(*l2, pos)
							}
						}
					}
				}
			}
			*l1 = *l2
			*l2 = (*l2)[:0]
		}
		for _, vis := range visited {
			if !vis {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func AvailableToMove(board *Board, piece *Piece) []Position {
	positions := []Position{}

	if CanMove(board, piece) {
		switch piece.Type {
		case QueenBee:
			// fmt.Println("QueenBee")
			for _, p := range board.Pieces {
				if p.Position.X != piece.Position.X || p.Position.Y != piece.Position.Y {
					for i := -1; i <= 1; i++ {
						for j := -1; j <= 1; j++ {
							if (i != 0 || j != 0) && i+j != 0 {
								pos := Position{X: piece.Position.X + i, Y: piece.Position.Y + j}
								if IsPositionNeignbour(pos, p.Position) && CanSqueezeThrough(board, piece.Position, pos) {
									cellFree := true
									for _, pp := range board.Pieces {
										if pp.Position.X == pos.X && pp.Position.Y == pos.Y {
											cellFree = false
											break
										}
									}
									if cellFree {
										positions = append(positions, pos)
									}
								}
							}
						}
					}
				}
			}
		case SoldierAnt:
			l1 := &[]Position{}
			l2 := &[]Position{}
			set := map[Position]bool{}
			for _, pp := range board.Pieces {
				set[pp.Position] = true
			}

			*l1 = append(*l1, piece.Position)

			for len(*l1) > 0 {
				for _, p := range *l1 {
					for i := -1; i <= 1; i++ {
						for j := -1; j <= 1; j++ {
							if (i != 0 || j != 0) && i+j != 0 {
								pos := Position{X: p.X + i, Y: p.Y + j}
								for _, pp := range board.Pieces {
									if (pp.Position.X != pos.X || pp.Position.Y != pos.Y) &&
										IsPositionNeignbour(pos, pp.Position) && CanSqueezeThrough(board, p, pos) {
										if _, ok := set[pos]; !ok {
											positions = append(positions, pos)
											*l2 = append(*l2, pos)
											set[pos] = true
										}
									}
								}
							}
						}
					}
				}
				*l1 = *l2
				*l2 = (*l2)[:0]
			}
		case Spider:
			l1 := &[]Position{}
			l2 := &[]Position{}
			set := map[Position]bool{}
			for _, pp := range board.Pieces {
				set[pp.Position] = true
			}

			*l1 = append(*l1, piece.Position)

			k := 0
			for len(*l1) > 0 && k < 3 {
				for _, p := range *l1 {
					for i := -1; i <= 1; i++ {
						for j := -1; j <= 1; j++ {
							if (i != 0 || j != 0) && i+j != 0 {
								pos := Position{X: p.X + i, Y: p.Y + j}
								for _, pp := range board.Pieces {
									if (pp.Position.X != pos.X || pp.Position.Y != pos.Y) &&
										IsPositionNeignbour(pos, pp.Position) && CanSqueezeThrough(board, p, pos) {
										if _, ok := set[pos]; !ok {
											if k == 2 {
												positions = append(positions, pos)
											}
											*l2 = append(*l2, pos)
											set[pos] = true
										}
									}
								}
							}
						}
					}
				}
				*l1 = *l2
				*l2 = (*l2)[:0]
				k += 1
			}
		case Grasshopper:
			set := map[Position]bool{}
			for _, pp := range board.Pieces {
				set[pp.Position] = true
			}

			for _, p := range board.Pieces {
				if IsPositionNeignbour(piece.Position, p.Position) {
					x := p.Position.X - piece.Position.X
					y := p.Position.Y - piece.Position.Y
					pos := Position{p.Position.X + x, p.Position.Y + y}
					_, ok := set[pos]
					for ok {
						pos.X += x
						pos.Y += y
						_, ok = set[pos]
					}
					positions = append(positions, pos)
				}
			}
		case Beetle:
			for _, p := range board.Pieces {
				if p.Position.X != piece.Position.X || p.Position.Y != piece.Position.Y {
					for i := -1; i <= 1; i++ {
						for j := -1; j <= 1; j++ {
							if (i != 0 || j != 0) && i+j != 0 {
								pos := Position{X: piece.Position.X + i, Y: piece.Position.Y + j}
								if IsPositionNeignbour(pos, p.Position) && CanSqueezeThrough(board, piece.Position, pos) {
									positions = append(positions, pos)
								}
							}
						}
					}
				}
			}
		}
	}
	return positions
}
