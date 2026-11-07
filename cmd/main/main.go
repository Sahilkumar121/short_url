package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sahilkumar121/short_url/internal/config"
	"github.com/Sahilkumar121/short_url/internal/database"
	"github.com/Sahilkumar121/short_url/internal/email"
	"github.com/Sahilkumar121/short_url/internal/handler"
	"github.com/Sahilkumar121/short_url/internal/middleware"
	"github.com/Sahilkumar121/short_url/internal/repo"
	"github.com/Sahilkumar121/short_url/internal/service"
	"github.com/Sahilkumar121/short_url/internal/worker"
)

func main() {

	// logger set
	loggerHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelError,
		AddSource: true,
	})
	logger := slog.New(loggerHandler)
	slog.SetDefault(logger)

	// config load
	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("error occur %v", err)
	}

	// database connection
	pgDb, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("error occur %v", err)
	}

	defer pgDb.Close()

	// smtp hanlder
	emailSender := email.NewSMTPHandler(cfg)

	// email worker handler
	emailWorker := worker.NewEmailWorker(emailSender, logger)
	if err := emailWorker.Start(); err != nil {
		log.Fatalf("error starting email worker %v", err)
	}
	defer emailWorker.Close()

	mux := http.NewServeMux()

	// repository
	urlRepo := repo.NewURLRepositoryHandler(pgDb)
	userRepo := repo.NewUserRepositoryHandler(pgDb)

	// service
	urlService := service.NewURLService(urlRepo)
	userService := service.NewUserService(userRepo, emailWorker, logger)

	// api handler
	healthHandler := handler.NewHealthHandler(logger)
	urlHandler := handler.NewUrlHandler(logger, urlService)
	userHandler := handler.NewUserHandler(logger, userService, cfg)

	// api
	mux.HandleFunc("GET /api/v1/healthz", healthHandler.GetHealth)
	mux.HandleFunc("GET /api/v1/url/longurl/{short_code}", urlHandler.GetLongUrl)
	mux.HandleFunc("POST /api/v1/user/login", userHandler.LoginUser)
	mux.HandleFunc("POST /api/v1/user/register", userHandler.RegisterUser)

	mux.Handle("POST /api/v1/url/shortcode", middleware.OptionalAuthMiddleware([]byte(cfg.SecretKey), http.HandlerFunc(urlHandler.GetShortCode)))
	mux.Handle("GET /api/v1/user/me", middleware.AuthMiddleware([]byte(cfg.SecretKey), http.HandlerFunc(userHandler.GetUserDetail)))

	requestIDmiddleware := middleware.RequestIdMiddleware(mux)
	server := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      requestIDmiddleware,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "err", err)
		}
	}()

	logger.Info("server is running", "port", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "err", err)
	}
}
