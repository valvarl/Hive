package api

import (
	"context"
	"hive/pkg/game"
)

type GameRequest struct {
	// Graph build.Graph
}

type GameStarted struct {
	ID game.ID
	// MissingFiles []build.ID
}

type StatusUpdate struct {
	// JobFinished   *JobResult
	// BuildFailed   *BuildFailed
	// BuildFinished *BuildFinished
}

type SignalRequest struct {
	// UploadDone *UploadDone
}

type SignalResponse struct {
}

type StatusWriter interface {
	Started(rsp *GameStarted) error
	Updated(update *StatusUpdate) error
}

type Service interface {
	StartGame(ctx context.Context, request *GameRequest, w StatusWriter) error
	SignalGame(ctx context.Context, buildID game.ID, signal *SignalRequest) (*SignalResponse, error)
}

type StatusReader interface {
	Close() error
	Next() (*StatusUpdate, error)
}
