package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Cerebrovinny/fizz-buzz-rest/internal/handler"
	"github.com/Cerebrovinny/fizz-buzz-rest/internal/middleware"
	"github.com/Cerebrovinny/fizz-buzz-rest/internal/statistics"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := statistics.NewStore()

	router := chi.NewRouter()

	h := handler.NewHandler(store)
	router.With(middleware.Statistics(store)).Get("/fizzbuzz", h.FizzBuzz)
	router.Get("/statistics", h.Statistics)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("received signal %s, shutting down server...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
