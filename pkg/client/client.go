package client

import (
	"context"

	"hive/pkg/api"
	"hive/pkg/game"

	"go.uber.org/zap"
)

type Client struct {
	log            *zap.Logger
	api            *api.GameClient
	engine         Engine
	engineStarted  bool
	engineResponse chan *game.Move
}

type Engine interface {
	Start(ctx context.Context, board *game.Board, hand, opponentHand *game.Hand, engineResponse chan *game.Move)
	Update(board *game.Board, hand, opponentHand *game.Hand)
}

func NewClient(l *zap.Logger, apiEndpoint string, engine Engine) *Client {
	client := &Client{
		log:            l,
		engine:         engine,
		engineStarted:  false,
		engineResponse: make(chan *game.Move, 1),
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
	if !c.engineStarted {
		go func() {
			c.engine.Start(ctx, su.GameState.Board, su.GameState.Hand, su.GameState.OpponentHand, c.engineResponse)
		}()
		c.engineStarted = true
	} else {
		c.engine.Update(su.GameState.Board, su.GameState.Hand, su.GameState.OpponentHand)
	}
	// ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	select {
	case <-ctx.Done():
		return nil
	case move := <-c.engineResponse:
		err := c.api.SendMove(api.PlayMove{GameID: su.GameID, Move: move})
		c.log.Info("Ход отправлен:", zap.Any("move", move))
		if err != nil {
			return err
		}
	}
	return nil
}
