package estest

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/special-case" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a special case"}`)) //nolint:errcheck // just a test case
		return
	}

	// Default case: serve files from the directory
	fileToReturn := filepath.Join("./test_data", r.URL.Path)

	data, err := os.ReadFile(fileToReturn)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data) //nolint:errcheck // just a test case
}

func StartServer() *http.Server {
	server := &http.Server{Addr: ":7104", Handler: http.HandlerFunc(mockHandler)} //nolint:gosec // just a test case

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("API server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	return server
}
