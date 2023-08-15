package client

import (
	"context"
	"time"

	"hive/pkg/api"
	"hive/pkg/game"

	"go.uber.org/zap"
)

type Client struct {
	log    *zap.Logger
	api    *api.GameClient
	engine Engine
}

type Engine interface {
	MakeMove(ctx context.Context, board *game.Board, hand, opponentHand *game.Hand) *game.Move
}

func NewClient(l *zap.Logger, apiEndpoint string, engine Engine) *Client {
	client := &Client{
		log:    l,
		engine: engine,
	}

	client.api = api.NewGameClient(l, apiEndpoint, client)
	return client
}

func (c *Client) Start(ctx context.Context) {
	err := c.api.Connect()
	if err != nil {
		c.log.Error("Ошибка присоединения к игре", zap.Error(err))
	}
	c.log.Info("Успешное присоединение к игре")

	err = c.api.HandleUpdates(ctx)
	if err != nil {
		c.log.Error("Ошибка игровой сессии", zap.Error(err))
	}

	err = c.api.Close()
	if err != nil {
		c.log.Error("Ошибка завершения подключения", zap.Error(err))
	}
	c.log.Info("Успешное завершение игры")
}

func (c *Client) HandleStatusUpdate(ctx context.Context, su *api.StatusUpdate) error {
	engineResponse := make(chan *game.Move)
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	go func() {
		engineResponse <- c.engine.MakeMove(ctx, su.GameState.Board, su.GameState.Hand, su.GameState.OpponentHand)
	}()
	select {
	case <-ctx.Done():
	case move := <-engineResponse:
		err := c.api.SendMove(api.PlayMove{GameID: su.GameID, Move: move})
		if err != nil {
			return err
		}
	}
	return nil
}
