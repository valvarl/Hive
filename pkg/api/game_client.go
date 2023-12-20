package api

import (
	"context"
	"encoding/json"
	"hive/pkg/game"
	"net"

	"go.uber.org/zap"
)

type GameClient struct {
	ID       game.ID
	logger   *zap.Logger
	endpoint string
	conn     net.Conn
	cs       ClientServise
}

func NewGameClient(logger *zap.Logger, endpoint string, cs ClientServise) *GameClient {
	return &GameClient{
		ID:       game.NewID(),
		logger:   logger,
		endpoint: endpoint,
		cs:       cs,
	}
}

func (c *GameClient) Connect() error {
	conn, err := net.Dial("tcp", c.endpoint)
	if err != nil {
		return err
	}
	c.conn = conn
	if err = c.Handshake(); err != nil {
		return err
	}
	return nil
}

func (c *GameClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *GameClient) HandleUpdates(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			statusUpdate, err := c.ReceiveStatusUpdate()
			if err != nil {
				return err
			}
			err = c.cs.HandleStatusUpdate(ctx, statusUpdate)
			if err != nil {
				return err
			}
		}
	}
}

func (c *GameClient) Handshake() error {
	handshake := Hanshake{PlayerID: c.ID}
	data, err := json.Marshal(handshake)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *GameClient) SendMove(move PlayMove) error {
	moveData, err := json.Marshal(move)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(moveData)
	if err != nil {
		return err
	}

	return nil
}

func (c *GameClient) ReceiveStatusUpdate() (*StatusUpdate, error) {
	var buffer []byte
	tempBuffer := make([]byte, 1024)
	// <-time.After(time.Second)
	for {
		n, err := c.conn.Read(tempBuffer)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, tempBuffer[:n]...)
		if n < 1024 {
			break // Прочитаны все данные
		}
	}

	var su StatusUpdate
	err := json.Unmarshal(buffer, &su)
	if err != nil {
		return nil, err
	}

	return &su, nil
}
