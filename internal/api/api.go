package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sts-issuer/internal/envs"
	"sts-issuer/internal/sts"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Start() {
	// The HTTP Server
	server := &http.Server{Addr: ":" + envs.GetEnvOrDefault("STS_PORTS", "3333"), Handler: service()}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/v1/issue", func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("id")
		if identifier == "" {
			http.Error(w, "Missing identifier in query", http.StatusBadRequest)
			return
		}

		creds, err := sts.GetCreds(identifier)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the STS credentials as JSON
		jsonData, err := json.Marshal(creds)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	})

	r.Get("/v1/list", func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("id")

		if identifier == "" {
			// No identifier provided, return all available STS data
			allSTSData := sts.GetAllSTSData()

			// Convert the result to JSON
			jsonData, err := json.Marshal(allSTSData)
			if err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
				return
			}

			// Return the list of all STS data
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
			return
		}

		// Call the function to get the STS data based on the identifier
		stsData, err := sts.GetSTSData(identifier)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert the result to JSON
		jsonData, err := json.Marshal(stsData)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}

		// Return the JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	return r
}
