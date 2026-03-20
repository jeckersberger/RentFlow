package main

import (
	"fmt"
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/config"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

func main() {
	cfg := config.Load("auth-service")
	log := logger.New(cfg.LogLevel, "auth-service")

	log.Info("Starting Auth Service", "port", cfg.ServicePort, "env", cfg.Environment)

	// TODO: Initialize repositories, use cases, handlers

	router := http.NewServeMux()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Ready endpoint
	router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	addr := fmt.Sprintf(":%d", cfg.ServicePort)
	log.Info("Listening on", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Server error", err)
	}
}
