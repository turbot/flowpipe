package estest

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// For test case: TestBasicAuth
func basicAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Define expected username and password
	expectedUsername := "testuser"
	expectedPassword := "testpass"

	// Extract the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// If no authorization header, send a WWW-Authenticate challenge
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode the provided credentials
	authEncoded := authHeader[len("Basic "):] // Strip "Basic " prefix
	authDecoded, err := base64.StdEncoding.DecodeString(authEncoded)
	if err != nil {
		http.Error(w, "Unauthorized - bad encoding", http.StatusUnauthorized)
		return
	}

	// Convert credentials from `username:password` to separate variables
	credentials := string(authDecoded)
	var username, password string
	colonIndex := len(credentials) - len(":") - len(expectedPassword) // approximate position of colon
	if colonIndex > 0 && colonIndex < len(credentials) {
		username = credentials[:colonIndex]
		password = credentials[colonIndex+1:]
	}

	// Validate credentials
	if username == expectedUsername && password == expectedPassword {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Authenticated successfully")
	} else {
		http.Error(w, "Unauthorized - invalid credentials", http.StatusUnauthorized)
	}
}

func mockHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/basic-auth-01" {
		basicAuthHandler(w, r)
		return
	}

	if r.URL.Path == "/special-case" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a special case"}`)) //nolint:errcheck // just a test case
		return
	}

	if r.URL.Path == "/delay" {
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "delay"}`)) //nolint:errcheck // just a test case
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
			slog.Error("Mock HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	return server
}
