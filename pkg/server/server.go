package server

import (
	"context"
	"errors"
	"fmt"
	"hive/pkg/api"
	"hive/pkg/game"
	"math/rand"

	"go.uber.org/zap"
)

type Server struct {
	log *zap.Logger
	api *api.GameServer
}

func NewServer(l *zap.Logger, endpoint string) *Server {
	server := &Server{
		log: l,
	}
	server.api = api.NewGameServer(l, endpoint, server)
	return server
}

func (s *Server) Start(ctx context.Context) {
	s.api.Start(ctx)
}

func (s *Server) CreateNewGame(first, second *api.Player) *api.Game {
	if rand.Float32() < 0.5 {
		first, second = second, first
	}

	game := &api.Game{
		ID:      game.NewID(),
		Players: []game.ID{first.ID, second.ID},
		Session: game.NewGameSession(game.StandardHand),
	}

	return game
}

func (s *Server) StartGame(ctx context.Context, game *api.Game) error {
	fp, err := s.api.GetPlayer(game.Players[0])
	if err != nil {
		return err
	}
	sp, err := s.api.GetPlayer(game.Players[1])
	if err != nil {
		return err
	}
	players := []*api.Player{fp, sp}
	su := &api.StatusUpdate{
		GameID:       game.ID,
		GameState:    &api.GameState{Board: game.Session.GetBoard(), Hand: game.Session.GetWhiteHand(), OpponentHand: game.Session.GetBlackHand()},
		GameFailed:   nil,
		GameFinished: nil,
	}
	for !game.Session.IsGameOver() {
		for i := 0; i <= 1; i++ {
			select {
			case <-ctx.Done():
				return nil
			default:
				err = s.api.SendStatusUpdate(players[i], su)
				if err != nil {
					s.log.Error("Ошибка при отправке статуса игроку", zap.Error(err))
					return err
				}

				if su.GameFailed != nil {
					s.log.Info(fmt.Sprintf("Ошибка сессии: %s", su.GameFailed.Error))
				}

				if su.GameFinished != nil {
					s.log.Info("Игра завершена", zap.Any("id", su.GameID),
						zap.Bool("tie", su.GameFinished.Tie),
						zap.Any("winner", su.GameFinished.Winer),
					)
				}

				move, err := s.api.ReceiveMove(players[i])
				if err != nil {
					s.log.Error("Ошибка при получении хода от игрока", zap.Error(err))
					return err
				}
				// TODO: here server can receive move from different game of same player
				su, err = s.UpdateGameState(game, move.Move)
				if err != nil {
					s.log.Error(err.Error())
				}
				su.GameID = move.GameID
				game.Session.NextTurn()
				su.GameState.Turn = game.Session.GetTurn()
			}
		}
	}
	return nil
}

func (s *Server) UpdateGameState(g *api.Game, move *game.Move) (*api.StatusUpdate, error) {
	// TODO: Check if move correct'
	var su *api.StatusUpdate
	if !move.Piece.Placed {
		color := game.White
		if !g.Session.WhiteToMove() {
			color = game.Black
		}

		availablePositions := game.AvailableToPlace(g.Session.GetBoard(), color)
		positionLegal := false

		for _, position := range availablePositions {
			if move.Position.X == position.X && move.Position.Y == position.Y {
				positionLegal = true
				break
			}
		}

		if positionLegal {
			var hand *game.Hand
			if g.Session.WhiteToMove() {
				hand = g.Session.GetWhiteHand()
			} else {
				hand = g.Session.GetBlackHand()
			}

			if hand.Pieces[move.Piece.Type] > 0 {
				if g.Session.GetTurn() >= 6 && hand.Pieces[game.QueenBee] != 0 {
					if move.Piece.Type != game.QueenBee {
						return nil, errors.New("королеву улья нужно разместить за 4 хода")
					}
				}
				piece := move.Piece
				piece.Placed = true
				piece.Position = *move.Position
				g.Session.GetBoard().Pieces = append(g.Session.GetBoard().Pieces, move.Piece)
				hand.Pieces[move.Piece.Type] -= 1
			} else {
				return nil, fmt.Errorf("недостаточно насекомых типа %T", move.Piece.Type)
			}
		}
	} else {
		availablePositions := game.AvailableToMove(g.Session.GetBoard(), move.Piece)
		positionLegal := false
		for _, pos := range availablePositions {
			if move.Position.X == pos.X && move.Position.Y == pos.Y {
				positionLegal = true
				break
			}
		}

		if positionLegal {
			var hand *game.Hand
			if g.Session.WhiteToMove() {
				hand = g.Session.GetWhiteHand()
			} else {
				hand = g.Session.GetBlackHand()
			}

			if g.Session.GetTurn() >= 6 && hand.Pieces[game.QueenBee] != 0 {
				return nil, errors.New("королеву улья нужно разместить за 4 хода")
			}

			s.log.Info("", zap.Any("board", g.Session.GetBoard().Pieces))

			for _, pos := range g.Session.GetBoard().Pieces {
				if move.Piece.Position.X == pos.Position.X && move.Piece.Position.Y == pos.Position.Y {
					s.log.Info("HELLO")
					(*pos).Position = game.Position{X: move.Position.X, Y: move.Position.Y}
					break
				}
			}

			s.log.Info("", zap.Any("board", g.Session.GetBoard().Pieces))
		}
	}
	su = &api.StatusUpdate{GameState: &api.GameState{}}
	su.GameState.Board = g.Session.GetBoard()
	if g.Session.WhiteToMove() {
		su.GameState.Hand = g.Session.GetBlackHand()
		su.GameState.OpponentHand = g.Session.GetWhiteHand()
	} else {
		su.GameState.Hand = g.Session.GetWhiteHand()
		su.GameState.OpponentHand = g.Session.GetBlackHand()
	}
	return su, nil
}
