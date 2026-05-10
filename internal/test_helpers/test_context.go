package testhelpers

import (
	"context"
	"errors"
	"finance_tracker/internal/config"
	"finance_tracker/internal/logger"
	"finance_tracker/internal/repository"
	"finance_tracker/internal/service/currency"
	"finance_tracker/internal/service/user"
	"finance_tracker/internal/storage/database"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type TestContainer struct {
	Ctx    context.Context
	Cfg    *config.AppConfig
	Logger logger.AppLogger

	DB   database.DBConnector
	Repo *repository.Repo

	ServiceUser     *user.Service
	ServiceCurrency *currency.Service
}

func GetClean(t *testing.T) *TestContainer {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	conf := getTestConfig()
	prepareTestDB(ctx, t, &conf.ConfigDB)

	dbConnect, err := database.InitDBConnect(ctx, &conf.ConfigDB, guessMigrationDir(t))
	require.NoError(t, err)
	cleanupDB(t, dbConnect)
	t.Cleanup(func() {
		cancel()
		require.NoError(t, dbConnect.Client().Close())
	})

	appLog := logger.NewAppSLogger()
	// repo init
	repo := repository.InitRepo(dbConnect)

	// service init
	srvCurrency := currency.NewService(ctx, appLog, conf, repo)
	serviceSampler := user.InitService(ctx, appLog, repo, conf)
	return &TestContainer{
		Ctx:             ctx,
		Cfg:             conf,
		Logger:          appLog,
		DB:              dbConnect,
		Repo:            repo,
		ServiceUser:     serviceSampler,
		ServiceCurrency: srvCurrency,
	}
}

func prepareTestDB(ctx context.Context, t *testing.T, cnf *config.DBConf) {
	dbConnect, err := database.InitDBConnect(ctx, &config.DBConf{
		Address:        cnf.Address,
		Port:           cnf.Port,
		User:           cnf.User,
		Pass:           cnf.Pass,
		DBName:         "postgres",
		MaxConnections: cnf.MaxConnections,
	}, "")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, dbConnect.Client().Close())
	}()
	if _, err = dbConnect.Client().Exec(fmt.Sprintf("CREATE DATABASE %s", cnf.DBName)); !isDatabaseExists(err) {
		require.NoError(t, err)
	}
}

func getTestConfig() *config.AppConfig {
	return &config.AppConfig{
		AppPort: 0,
		ConfigDB: config.DBConf{
			Address:        "localhost",
			Port:           "5449",
			User:           "aHAjeK",
			Pass:           "AOifjwelmc8dw",
			DBName:         "sybill_test",
			MaxConnections: 10,
		},
		ExchangeToken: os.Getenv("EXCHANGE_TOKEN"),
		Auth: config.AuthConf{
			UIBaseURL: "http://localhost:3000",
			Token: config.TokenSigningConf{
				SigningKey:        "test-signing-key",
				ExpirationMinutes: 60,
			},
			Google: config.GoogleOAuth2Conf{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURL:  "http://localhost:8000/api/auth/google/callback",
			},
		},
	}
}

func isDatabaseExists(err error) bool {
	return checkSQLError(err, "42P04")
}

func checkSQLError(err error, code string) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	ok := errors.As(err, &pqErr)
	if !ok {
		return false
	}
	return string(pqErr.Code) == code
}

func guessMigrationDir(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	res := strings.Split(dir, "/internal")
	return res[0] + "/migrations"
}

func cleanupDB(t *testing.T, connector database.DBConnector) {
	for _, table := range repository.AllTables {
		_, err := connector.Client().Exec(fmt.Sprintf("TRUNCATE %s CASCADE", table))
		require.NoError(t, err)
	}
}
