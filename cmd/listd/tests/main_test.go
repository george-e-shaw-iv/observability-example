// Package tests contain integration tests for the API provided by the list daemon.
package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/george-e-shaw-iv/observability-example/cmd/listd/handlers"
	"github.com/george-e-shaw-iv/observability-example/internal/platform/testdb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// a is a reference to the main Application type. This is used for its database
// connection that it harbours inside of the type as well as the route definitions
// that are defined on the embedded handler.
var a *handlers.Application

// TestMain calls testMain and passes the returned exit code to os.Exit(). The reason
// that TestMain is basically a wrapper around testMain is because os.Exit() does not
// respect deferred functions, so this configuration allows for a deferred function.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// testMain returns an integer denoting an exit code to be returned and used in
// TestMain. The exit code 0 denotes success, all other codes denote failure (1
// and 2).
func testMain(m *testing.M) int {
	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := zCfg.Build()
	if err != nil {
		fmt.Printf("error initializing logger: %v\n", err)
		return 1
	}

	dbc, err := testdb.Open(logger)
	if err != nil {
		logger.Error("create test database connection", zap.Error(err))
		return 1
	}
	defer dbc.Close()

	a = handlers.NewApplication(dbc, logger)

	return m.Run()
}
