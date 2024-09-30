package main

import (
	"golang.org/x/exp/slog"
	"os"
	"studentgit.kata.academy/zidame/go-kata/test/url-shortener/internal/config"
)

const (
	envlocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	//fmt.Printf("%+v\n", cfg)
	log := setupLogger(cfg.Env)

	log.Info("starting server", slog.String("env", cfg.Env))
	log.Debug("debug massenges are enable")
}

func setupLogger(env string) *slog.Logger {
	// kjuth
	var log *slog.Logger
	switch env {
	case envlocal:
		//функция которая создает новый логер
		// она принимает параметр обработчика логов  который управляет тем куда  как и куда
		// буду выполняяться логи
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
