package main

import (
	"net/http"
)

func main() {
	// Step 1: Create a new http.ServeMux
	mux := http.NewServeMux()

	// Step 2: Wrap that mux in a custom middleware function that adds CORS headers to the response
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
