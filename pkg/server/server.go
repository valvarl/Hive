package server

import (
	"context"
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
		GameState:    &api.GameState{Board: game.Session.GetBoard(), Hand: game.Session.GetWhiteHand(), OpponentHand: game.Session.GetWhiteHand()},
		GameFailed:   nil,
		GameFinished: nil,
	}
	for !game.Session.IsGameOver() {
		for i := 0; i <= 1; i++ {
			select {
			case <-ctx.Done():
				return nil
			default:
				err = s.api.SendStatusUpdate(players[(i+1)%2], su)
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

				su = s.UpdateGameState(move.Move)
				su.GameID = move.GameID
			}
		}
	}
	return nil
}

func (s *Server) UpdateGameState(move *game.Move) *api.StatusUpdate {
	return &api.StatusUpdate{}
}
