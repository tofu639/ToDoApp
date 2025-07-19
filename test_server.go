package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Wait a moment for server to start
	time.Sleep(2 * time.Second)
	
	// Test health check endpoint
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("Health check status: %d\n", resp.StatusCode)
}