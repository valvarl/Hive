package api

import (
	"context"
	"encoding/json"
	"hive/pkg/game"
	"net"
	"time"

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
	buffer := make([]byte, 1024)
	<-time.After(time.Second)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	var su StatusUpdate
	err = json.Unmarshal(buffer[:n], &su)
	if err != nil {
		return nil, err
	}

	return &su, nil
}

// func main() {
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatal("Ошибка при инициализации логгера:", err)
// 	}

// 	// Создание экземпляра игрока
// 	player, err := NewPlayerClient(logger, "localhost:8080")
// 	if err != nil {
// 		logger.Fatal("Ошибка при создании игрока:", zap.Error(err))
// 	}

// 	// Подключение к игре
// 	err = player.Connect()
// 	if err != nil {
// 		logger.Fatal("Ошибка при подключении к игре:", zap.Error(err))
// 	}

// 	// Запуск горутины для приема обновлений от сервера
// 	go player.ReceiveUpdates()

// 	// Пример отправки хода на сервер
// 	move := Move{
// 		Piece: &game.Piece{
// 			// Задайте поля вашего хода здесь
// 		},
// 		Position: &game.Position{
// 			// Задайте поля вашего хода здесь
// 		},
// 	}

// 	err = player.PlayMove(move)
// 	if err != nil {
// 		logger.Error("Ошибка при отправке хода:", zap.Error(err))
// 	}

// 	// Задержка для получения обновлений от сервера
// 	time.Sleep(5 * time.Second)

// 	// Завершение программы
// 	logger.Sync()
// }
