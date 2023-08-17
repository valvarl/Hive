package api

import (
	"context"
	"hive/pkg/game"
)

type Hanshake struct {
	PlayerID game.ID
}

type PlayMove struct {
	GameID game.ID
	Move   *game.Move
}

type StatusUpdate struct {
	GameID       game.ID
	GameState    *GameState
	GameFailed   *GameFailed
	GameFinished *GameFinished
}

type GameFailed struct {
	Error string
}

type GameFinished struct {
	Winer bool
	Tie   bool
}

type GameState struct {
	Board        *game.Board
	Hand         *game.Hand
	OpponentHand *game.Hand
}

type ClientServise interface {
	HandleStatusUpdate(ctx context.Context, statusUpdate *StatusUpdate) error
}

type ServerServise interface {
	CreateNewGame(first, second *Player) *Game
	StartGame(ctx context.Context, game *Game) error
	UpdateGameState(game *Game, move *game.Move) *StatusUpdate
}
