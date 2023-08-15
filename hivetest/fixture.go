package hivetest

import (
	"context"
	"errors"
	"hive/pkg/client"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type env struct {
	RootDir string
	Logger  *zap.Logger

	Ctx context.Context

	Clients []*client.Client
}

const (
	logToStderr = true
)

type Config struct {
	WorkerCount int
}

func newEnv(t *testing.T, config *Config) (e *env, cancel func()) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	absCWD, err := filepath.Abs(cwd)
	require.NoError(t, err)

	rootDir := filepath.Join(absCWD, "workdir", t.Name())
	require.NoError(t, os.RemoveAll(rootDir))

	if err = os.MkdirAll(rootDir, 0777); err != nil {
		if errors.Is(err, os.ErrPermission) {
			rootDir, err = ioutil.TempDir("", "")
			require.NoError(t, err)
		} else {
			require.NoError(t, err)
		}
	}

	env := &env{
		RootDir: rootDir,
	}

	cfg := zap.NewDevelopmentConfig()

	if runtime.GOOS == "windows" {
		cfg.OutputPaths = []string{filepath.Join("winfile://", env.RootDir, "test.log")}
		err = zap.RegisterSink("winfile", newWinFileSink)
		require.NoError(t, err)
	} else {
		cfg.OutputPaths = []string{filepath.Join(env.RootDir, "test.log")}
	}

	if logToStderr {
		cfg.OutputPaths = append(cfg.OutputPaths, "stderr")
	}

	env.Logger, err = cfg.Build()
	require.NoError(t, err)

	t.Helper()
	t.Logf("test is running inside %s; see test.log file for more info", filepath.Join("workdir", t.Name()))

	// port, err := testtool.GetFreePort()
	require.NoError(t, err)
	// addr := "127.0.0.1:" + port
	// serverEndpoint := "http://" + addr + "/server"

	// var cancelRootContext func()
	// env.Ctx, cancelRootContext = context.WithCancel(context.Background())

	// env.Client = client.NewClient(
	// 	env.Logger.Named("client"),
	// 	serverEndpoint,
	// 	filepath.Join(absCWD, "testdata", t.Name()))

	return nil, nil
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	if len(u.Opaque) > 0 {
		// Remove leading slash left by url.Parse()
		return os.OpenFile(u.Opaque[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	}
	// if url.URL is empty, don't panic slice index error
	return os.OpenFile(u.Opaque, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
