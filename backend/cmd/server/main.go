package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/api"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/auth"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/cache"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/config"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/mqtt"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/store"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbCtx, cancelDB := context.WithTimeout(ctx, 15*time.Second)
	defer cancelDB()

	pgStore, err := store.NewPostgresStore(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	if err := pgStore.RunMigrations(dbCtx, cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to run SQL migrations from %q: %v", cfg.MigrationsDir, err)
	}
	log.Printf("PostgreSQL + PostGIS ready (migrations applied from %s)", cfg.MigrationsDir)

	authenticator, err := auth.NewAuthenticator(
		cfg.AdminPassphraseHash,
		cfg.AdminPassphraseSalt,
		cfg.SessionSecret,
	)
	if err != nil {
		log.Fatalf("invalid authentication configuration: %v", err)
	}

	memCache := cache.NewMemoryCache(60 * time.Second)
	limiter := auth.NewIPRateLimiter(10, time.Minute)

	var mqttSub *mqtt.Subscriber
	if cfg.MQTTEnabled {
		mqttSub = mqtt.NewSubscriber(cfg.MQTTBrokerURL, pgStore, memCache)
		mqttSub.Start(ctx)
		defer mqttSub.Close()
		log.Printf("MQTT subscriber initialized for broker %s", cfg.MQTTBrokerURL)
	} else {
		log.Printf("MQTT subscriber disabled (MQTT_ENABLED=false)")
	}

	handler := api.NewHandler(
		pgStore,
		memCache,
		authenticator,
		limiter,
		cfg.Spatial,
		cfg.CookieSecure,
	)
	router := api.NewRouter(handler, cfg.CORSAllowedOrigins)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down server...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Printf("server stopped cleanly")
}
