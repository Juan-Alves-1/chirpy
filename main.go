package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) serveMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
		`, cfg.fileserverHits)))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset"))
}

func main() {
	// Step 1: Create a new http.ServeMux
	mux := http.NewServeMux()

	// Initialize your config struct
	apiCfg := &apiConfig{}

	// Create a file server which serves files out of the current directory
	// The file server uses the '.' (dot) to indicate the current directory
	fileServer := http.FileServer(http.Dir("."))
	// Use the Handle method to register the file server as the handler for the root path
	// Wrap the file server handler with your new middleware
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))

	// Setup file server for static assets
	// StripPrefix removes the "/assets" prefix before looking in the assets directory
	assetsFileServer := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", assetsFileServer))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.serveMetrics)
	mux.HandleFunc("GET /api/reset", apiCfg.resetHits)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	// Step 2: Wrap that mux in a custom middleware function that adds CORS headers to the response
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			// Respond to the CORS preflight request
			w.WriteHeader(http.StatusOK)
			return
		}
		// Call the original ServeMux with the request
		mux.ServeHTTP(w, r)
	})

	// Step 3: Create a new http.Server and use the corsMux as the handler
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: corsMux,
	}

	// Step 4: Use the server's ListenAndServe method to start the server
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Handle errors other than the expected graceful shutdown
		panic(err)
	}
}
