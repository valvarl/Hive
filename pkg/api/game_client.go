package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hive/pkg/game"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type GameClient struct {
	log      *zap.Logger
	endpoint string
	client   *http.Client
}

func NewBuildClient(l *zap.Logger, endpoint string) *GameClient {
	return &GameClient{
		log:      l,
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

func (c *GameClient) StartGame(ctx context.Context, request *GameRequest) (*GameStarted, StatusReader, error) {
	c.log.Info("Starting a new game", zap.Any("request", request))

	body, err := json.Marshal(request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/start", c.endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		errorMessage := res.Header.Get("X-Error-Message")
		return nil, nil, fmt.Errorf("unexpected status code: %d, message: %s", res.StatusCode, errorMessage)
	}

	decoder := json.NewDecoder(res.Body)

	var buildStarted GameStarted
	if err := decoder.Decode(&buildStarted); err != nil {
		defer res.Body.Close()
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &buildStarted, &statusReader{decoder, res.Body, res.Header}, nil
}

func (c *GameClient) SignalGame(ctx context.Context, gameID game.ID, signal *SignalRequest) (*SignalResponse, error) {
	c.log.Info("Sending signal to the game", zap.String("gameID", gameID.String()))

	body, err := json.Marshal(signal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/signal?game_id=%s", c.endpoint, gameID), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errorMessage := res.Header.Get("X-Error-Message")
		return nil, fmt.Errorf("unexpected status code: %d, message: %s", res.StatusCode, errorMessage)
	}

	var signalResponse SignalResponse
	if err := json.NewDecoder(res.Body).Decode(&signalResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &signalResponse, nil
}

type statusReader struct {
	decoder *json.Decoder
	body    io.ReadCloser
	headers http.Header
}

func (r *statusReader) Close() error {
	return r.body.Close()
}

func (r *statusReader) Next() (*StatusUpdate, error) {
	var update StatusUpdate
	if err := r.decoder.Decode(&update); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to decode status update: %w", err)
	}

	return &update, nil
}
