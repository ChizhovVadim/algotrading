package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ChizhovVadim/algotrading/internal/cli"
	"github.com/ChizhovVadim/algotrading/internal/traderapp"
)

var (
	gitRevision string
	buildDate   string
)

// Автоторговля торговых советников
func main() {
	config, err := loadConfig("trader.xml")
	if err != nil {
		panic(err)
	}

	logFolderPath, err := getLogFolder()
	if err != nil {
		panic(err)
	}

	var todayName = time.Now().Format("2006-01-02")

	logFile, err := appendFile(filepath.Join(logFolderPath, todayName+".txt"))
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	var logger = slog.New(cli.Fanout(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
		slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	))

	logger.Debug("Environment",
		"BuildDate", buildDate,
		"GitRevision", gitRevision)

	logger.Debug("runtime",
		"Version", runtime.Version(),
		"NumCPU", runtime.NumCPU(),
		"GOMAXPROCS", runtime.GOMAXPROCS(0))

	apiLogFile, err := appendFile(filepath.Join(logFolderPath, todayName+"-api.txt"))
	if err != nil {
		panic(err)
	}
	defer apiLogFile.Close()

	// Чтобы логировать запросы/ответы к внешним API без экранирования символов.
	var apiLogger = log.New(apiLogFile, "", log.LstdFlags|log.Lmicroseconds)

	var app = traderapp.New(logger, apiLogger, config)
	defer app.Close()

	err = app.Run()
	if err != nil {
		logger.Error("run failed",
			"error", err)
		return
	}
}

func getLogFolder() (string, error) {
	// TODO если запускать несколько экземпляров программы, то лучше задавать путь в конфиге
	var res = cli.MapPath("~/TradingData/Logs/luatrader/golang")
	if err := os.MkdirAll(res, os.ModePerm); err != nil {
		return "", err
	}
	return res, nil
}

func appendFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
