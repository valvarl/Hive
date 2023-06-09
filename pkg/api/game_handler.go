package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hive/pkg/game"
	"net"
	"sync"

	"go.uber.org/zap"
)

type Game struct {
	ID      game.ID
	Players []game.ID
	Session *game.GameSession
}

type GameServer struct {
	log      *zap.Logger
	endpoint string

	gameMu sync.Mutex
	games  map[game.ID]*Game

	playerMu sync.Mutex
	players  map[game.ID]*Player

	wg sync.WaitGroup
	ss ServerServise
}

func (gs *GameServer) AddGame(game *Game) {
	gs.gameMu.Lock()
	defer gs.gameMu.Unlock()
	gs.games[game.ID] = game
}

func (gs *GameServer) RemoveGame(ID game.ID) {
	gs.gameMu.Lock()
	defer gs.gameMu.Unlock()
	delete(gs.games, ID)
}

func (s *GameServer) GetActiveGameCount() int {
	s.gameMu.Lock()
	defer s.gameMu.Unlock()
	return len(s.games)
}

type Player struct {
	ID     game.ID
	conn   net.Conn
	gameMu sync.Mutex
	gameID map[game.ID]*Game
}

func (p *Player) AddGame(game *Game) {
	p.gameMu.Lock()
	defer p.gameMu.Unlock()
	p.gameID[game.ID] = game
}

func (p *Player) RemoveGame(ID game.ID) {
	p.gameMu.Lock()
	defer p.gameMu.Unlock()
	delete(p.gameID, ID)
}

func NewGameServer(logger *zap.Logger, endpoint string) (*GameServer, error) {
	return &GameServer{
		log:      logger,
		endpoint: endpoint,
		games:    make(map[game.ID]*Game),
		players:  make(map[game.ID]*Player),
	}, nil
}

func (s *GameServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return err
	}

	s.log.Info("Сервер запущен. Ожидание подключений...")

	var waitingPlayer *Player
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if s.GetActiveGameCount() >= 20 {
				continue
			}

			conn, err := listener.Accept()
			if err != nil {
				s.log.Error("Ошибка при принятии подключения:", zap.Error(err))
				continue
			}
			hs, err := s.Handshake(conn)
			if err != nil {
				s.log.Error("Ошибка аунтификации:", zap.Error(err))
				continue
			}

			var player *Player
			{
				s.playerMu.Lock()
				defer s.playerMu.Unlock()

				var ok bool
				player, ok = s.players[hs.PlayerID]
				if !ok {
					player = &Player{ID: hs.PlayerID, conn: conn, gameID: make(map[game.ID]*Game)}
					s.players[player.ID] = player
				}
			}

			if waitingPlayer == nil {
				waitingPlayer = player
				continue
			} else if waitingPlayer.ID == player.ID {
				continue
			}

			game := s.ss.CreateNewGame(waitingPlayer, player)
			waitingPlayer.AddGame(game)
			player.AddGame(game)
			s.AddGame(game)
			s.wg.Add(1)
			go func(fp, sp *Player) {
				if err = s.ss.StartGame(ctx, game); err != nil {
					s.log.Error("Ошибка игровой сессии", zap.Error(err))
				}
				fp.RemoveGame(game.ID)
				sp.RemoveGame(game.ID)
				s.RemoveGame(game.ID)
			}(waitingPlayer, player)

			s.log.Info("Игра началась. Игроки:", zap.Any("first", waitingPlayer.ID), zap.Any("second", player.ID))
			waitingPlayer = nil
		}
	}
}

func (s *GameServer) Handshake(conn net.Conn) (*Hanshake, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	var handshake Hanshake
	if err = json.Unmarshal(buffer[:n], &handshake); err != nil {
		return nil, err
	}

	return &handshake, nil
}

func (s *GameServer) GetPlayer(playerID game.ID) (*Player, error) {
	s.playerMu.Lock()
	defer s.playerMu.Unlock()

	player, ok := s.players[playerID]
	if !ok {
		return nil, fmt.Errorf("player ID not found: %v", playerID)
	}
	return player, nil
}

func (s *GameServer) ReceiveMove(player *Player) (*PlayMove, error) {
	buffer := make([]byte, 1024)
	n, err := player.conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	var move PlayMove
	err = json.Unmarshal(buffer[:n], &move)
	if err != nil {
		return nil, err
	}

	return &move, nil
}

func (s *GameServer) SendStatusUpdate(player *Player, su *StatusUpdate) error {
	data, err := json.Marshal(su)
	if err != nil {
		return err
	}

	_, err = player.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// func main() {
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		log.Fatal("Ошибка при инициализации логгера:", err)
// 	}

// 	server, err := NewGameServer(logger, "localhost:8080")
// 	if err != nil {
// 		logger.Fatal("Ошибка при создании сервера:", zap.Error(err))
// 	}

// 	ctx, _ := context.WithCancel(context.Background())
// 	err = server.Start(ctx)
// 	if err != nil {
// 		logger.Fatal("Ошибка при запуске сервера:", zap.Error(err))
// 	}

// 	logger.Sync()
// }
