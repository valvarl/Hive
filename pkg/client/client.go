package client

import (
	"context"

	"hive/pkg/api"
	"hive/pkg/game"

	"go.uber.org/zap"
)

type Client struct {
	log           *zap.Logger
	api           *api.GameClient
	engine        Engine
	engineStarted bool
}

type Engine interface {
	Start(ctx context.Context, board *game.Board, hand, opponentHand *game.Hand, engineResponse chan *game.Move)
}

func NewClient(l *zap.Logger, apiEndpoint string, engine Engine) *Client {
	client := &Client{
		log:           l,
		engine:        engine,
		engineStarted: false,
	}

	client.api = api.NewGameClient(l, apiEndpoint, client)
	return client
}

func (c *Client) Start(ctx context.Context) {
	err := c.api.Connect()
	if err != nil {
		c.log.Error("Ошибка присоединения к игре", zap.Error(err))
		return
	}
	c.log.Info("Успешное присоединение к игре", zap.String("ID", c.api.ID.String()))

	err = c.api.HandleUpdates(ctx)
	if err != nil {
		c.log.Error("Ошибка игровой сессии", zap.Error(err))
		return
	}

	err = c.api.Close()
	if err != nil {
		c.log.Error("Ошибка завершения подключения", zap.Error(err))
		return
	}
	c.log.Info("Успешное завершение игры")
}

func (c *Client) HandleStatusUpdate(ctx context.Context, su *api.StatusUpdate) error {
	engineResponse := make(chan *game.Move, 1)
	if !c.engineStarted {
		go func() {
			c.engine.Start(ctx, su.GameState.Board, su.GameState.Hand, su.GameState.OpponentHand, engineResponse)
		}()
		c.engineStarted = true
	}
	// ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case move := <-engineResponse:
			err := c.api.SendMove(api.PlayMove{GameID: su.GameID, Move: move})
			c.log.Info("Ход отправлен:", zap.Any("move", move))
			if err != nil {
				return err
			}
		}
	}
}
