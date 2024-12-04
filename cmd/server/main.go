package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rownd/client-go/pkg/rownd"
)

func main() {
	// Parse command line flags
	var (
		appKey    = flag.String("app-key", os.Getenv("ROWND_APP_KEY"), "Rownd app key")
		appSecret = flag.String("app-secret", os.Getenv("ROWND_APP_SECRET"), "Rownd app secret")
		addr      = flag.String("addr", ":3333", "HTTP server address")
	)
	flag.Parse()

	// Validate required flags
	if *appKey == "" || *appSecret == "" {
		log.Fatal("app-key and app-secret are required")
	}

	// Create client
	client, err := rownd.NewClient(
		rownd.WithAppKey(*appKey),
		rownd.WithAppSecret(*appSecret),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create server
	srv := setupServer(*addr, client)

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	log.Printf("Server started on %s", *addr)

	<-done
	log.Print("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Print("Server stopped")
}

func setupServer(addr string, client *rownd.Client) *http.Server {
	mux := http.NewServeMux()

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
