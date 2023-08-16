package hivetest

import (
	"testing"
)

var singleWorkerConfig = &Config{WorkerCount: 1}

func TestTwoPlayers(t *testing.T) {
	env, cancel := newEnv(t, singleWorkerConfig)
	defer cancel()

	defer env.Clients[0].Start(env.Ctx)
	go env.Clients[1].Start(env.Ctx)
}
