package main

import (
	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/interfaces"
	"L0/internal/kafka"
	"L0/internal/logger"
	"L0/internal/repository"
	"L0/internal/retry"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger.Init("order-service")
	logger.Info("Application starting")

	cfg := config.Load()
	logger.Info("Configuration loaded")

	var repo interfaces.OrderRepository
	var repoErr error

	ctx := context.Background()
	err := retry.Retry(ctx, "initialize_database", retry.RetryConfig{
		MaxAttempts: 5,
		InitialDelay: 2 * time.Second,
		MaxDelay: 30 * time.Second,
		Multiplier: 2.0,
	}, func() error {
		repo, repoErr = repository.NewPostgresRepository(&cfg.Database)
		return repoErr
	})

	if err != nil {
		logger.Error("Failed to initialize repository after retries", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	logger.Info("Database repository initialized")

	orderCache := cache.New()
	defer orderCache.Stop()
	logger.Info("Cache initialized")

	err = orderCache.LoadFromRepository(repo.GetAllOrders)
	if err != nil {
		logger.Warn("Failed to load cache from repository", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("Cache loaded from database", map[string]interface{}{
			"orders_count": orderCache.Count(),
		})
	}

	kafkaConsumer := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.GroupID,
		repo,
		orderCache,
	)
	defer kafkaConsumer.Close()
	logger.Info("Kafka consumer initialized", map[string]interface{}{
		"topic":    cfg.Kafka.Topic,
		"group_id": cfg.Kafka.GroupID,
		"brokers":  cfg.Kafka.Brokers,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("Starting Kafka consumer")
		kafkaConsumer.Start(ctx)
	}()

	http.Handle("/", http.FileServer(http.Dir("web/static")))

	http.HandleFunc("/order/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		orderHandler(w, r, orderCache, repo)
	}))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr: ":" + cfg.Server.Port,
	}

	go func() {
		logger.Info("Starting HTTP server", map[string]interface{}{
			"port": cfg.Server.Port,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", err)
			os.Exit(1)
		}
	}()

	waitForShutdown(server, cancel)
}

func waitForShutdown(server *http.Server, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", err)
	}

	logger.Info("Server stopped gracefully")
}

func orderHandler(w http.ResponseWriter, r *http.Request, cache interfaces.OrderCache, repo interfaces.OrderRepository) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderUID := r.URL.Path[len("/order/"):]
	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	if order, exists := cache.Get(orderUID); exists {
		writeJSONResponse(w, order)
		logger.Info("Order served from cache", map[string]interface{}{
			"order_uid": orderUID,
		})
		return
	}

	order, err := repo.GetOrderByUID(orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		logger.Warn("Order not found", map[string]interface{}{
			"order_uid": orderUID,
			"error":     err.Error(),
		})
		return
	}

	cache.Set(order)
	writeJSONResponse(w, order)
	logger.Info("Order served from database", map[string]interface{}{
		"order_uid": orderUID,
	})
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		logger.Error("JSON encoding error", err)
		return
	}

	w.Write(jsonData)
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}